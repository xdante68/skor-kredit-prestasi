package repo

import (
	"context"
	"errors"
	"fiber/skp/app/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type AchievementRepository interface {
	Create(studentID uuid.UUID, req model.CreateAchievementRequest) (*model.AchievementResponse, error)
	FindByAchievementID(id uuid.UUID) (*model.AchievementResponse, error)
	FindAll(role string, userID uuid.UUID, page, limit int, search, sortBy, order string) ([]model.AchievementResponse, int64, error)
	Update(id uuid.UUID, req model.UpdateAchievementRequest) (*model.AchievementResponse, error)
	UpdateStatus(id uuid.UUID, status string, verifierID *uuid.UUID, note string, points int) error
	Delete(id uuid.UUID) error
	AddAttachment(id uuid.UUID, attachment model.Attachment) error
	mapToResponse(ref model.AchievementReference, mongoDoc model.AchievementMongo) *model.AchievementResponse
	GetOwnerID(id uuid.UUID) (uuid.UUID, error)
	IsAdvisor(advisorID uuid.UUID, achievementID uuid.UUID) (bool, error)
	GetStatus(id uuid.UUID) (string, error)
	GetHistory(id uuid.UUID) (*model.AchievementHistoryResponse, error)
}

type AchievementRepo struct {
	pgDB    *gorm.DB
	mongoDB *mongo.Database
}

func NewAchievementRepo(pgDB *gorm.DB, mongoDB *mongo.Database) *AchievementRepo {
	return &AchievementRepo{pgDB: pgDB, mongoDB: mongoDB}
}

func (r *AchievementRepo) Create(studentID uuid.UUID, req model.CreateAchievementRequest) (*model.AchievementResponse, error) {
	var details model.AchievementDetails
	if req.CompetitionDetails != nil {
		details.CompetitionName = req.CompetitionDetails.CompetitionName
		details.CompetitionLevel = req.CompetitionDetails.CompetitionLevel
		details.Rank = req.CompetitionDetails.Rank
		details.MedalType = req.CompetitionDetails.MedalType
	} else if req.PublicationDetails != nil {
		details.PublicationTitle = req.PublicationDetails.PublicationTitle
		details.Authors = req.PublicationDetails.Authors
		details.Publisher = req.PublicationDetails.Publisher
		details.ISSN = req.PublicationDetails.ISSN
	} else if req.OrganizationDetails != nil {
		details.OrganizationName = req.OrganizationDetails.OrganizationName
		details.Position = req.OrganizationDetails.Position

		layout := "2006-01-02"
		start, err := time.Parse(layout, req.OrganizationDetails.StartDate)
		if err != nil {
			return nil, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD: %w", err)
		}
		end, err := time.Parse(layout, req.OrganizationDetails.EndDate)
		if err != nil {
			return nil, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD: %w", err)
		}
		details.Period = &model.OrganizationPeriod{
			Start: start,
			End:   end,
		}
	}

	now := time.Now()

	mongoData := bson.M{
		"studentId":       studentID.String(),
		"achievementType": req.AchievementType,
		"title":           req.Title,
		"description":     req.Description,
		"details":         details,
		"tags":            req.Tags,
		"attachments":     []interface{}{}, // Default kosong
		"points":          0,               // Default 0
		"createdAt":       now,
		"updatedAt":       now,
	}

	coll := r.mongoDB.Collection("achievements")
	res, err := coll.InsertOne(context.TODO(), mongoData)
	if err != nil {
		return nil, err
	}
	oid := res.InsertedID.(primitive.ObjectID)

	pgData := model.AchievementReference{
		StudentID:          studentID,
		MongoAchievementID: oid.Hex(),
		Status:             "draft",
	}

	if err := r.pgDB.Create(&pgData).Error; err != nil {
		coll.DeleteOne(context.TODO(), bson.M{"_id": oid})
		return nil, err
	}
	return &model.AchievementResponse{
		ID:              pgData.ID,
		MongoID:         oid.Hex(),
		StudentID:       studentID,
		Status:          "draft",
		AchievementType: req.AchievementType,
		Title:           req.Title,
		Description:     req.Description,
		Details:         details,
		Tags:            req.Tags,
		Attachments:     []model.Attachment{},
		Points:          0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (r *AchievementRepo) FindByAchievementID(id uuid.UUID) (*model.AchievementResponse, error) {
	var ref model.AchievementReference
	if err := r.pgDB.Preload("Student.User").Where("status != ?", model.StatusDeleted).First(&ref, "id = ?", id).Error; err != nil {
		return nil, err
	}

	var mongoDetail model.AchievementMongo
	objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)

	coll := r.mongoDB.Collection("achievements")
	err := coll.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&mongoDetail)

	if err != nil {
		return nil, errors.New("detail data missing in NoSQL")
	}

	return r.mapToResponse(ref, mongoDetail), nil
}

func (r *AchievementRepo) FindAll(role string, userID uuid.UUID, page, limit int, search, sortBy, order string) ([]model.AchievementResponse, int64, error) {
	var refs []model.AchievementReference
	var total int64

	query := r.pgDB.Preload("Student.User").Where("status != ?", model.StatusDeleted)

	if search != "" {
		coll := r.mongoDB.Collection("achievements")
		filter := bson.M{"title": bson.M{"$regex": search, "$options": "i"}}
		cursor, err := coll.Find(context.TODO(), filter)
		if err != nil {
			return nil, 0, err
		}
		defer cursor.Close(context.TODO())

		var mongoIDs []string
		for cursor.Next(context.TODO()) {
			var doc struct {
				ID primitive.ObjectID `bson:"_id"`
			}
			if err := cursor.Decode(&doc); err == nil {
				mongoIDs = append(mongoIDs, doc.ID.Hex())
			}
		}

		if len(mongoIDs) == 0 {
			return []model.AchievementResponse{}, 0, nil
		}
		query = query.Where("mongo_achievement_id IN ?", mongoIDs)
	}

	if role == model.RoleMahasiswa {
		var student model.Student
		if err := r.pgDB.Where("user_id = ?", userID).First(&student).Error; err != nil {
			return nil, 0, err
		}
		query = query.Where("student_id = ?", student.ID)

	} else if role == model.RoleDosenWali {
		var lecturer model.Lecturer
		if err := r.pgDB.Where("user_id = ?", userID).First(&lecturer).Error; err != nil {
			return nil, 0, err
		}
		query = query.Joins("JOIN students ON students.id = achievement_references.student_id").
			Where("students.advisor_id = ?", lecturer.ID).
			Where("achievement_references.status != ?", model.StatusDraft)
	}
	if role == model.RoleAdmin {
		query = query.Where("achievement_references.status != ?", model.StatusDraft)
	}

	if err := query.Model(&model.AchievementReference{}).Count(&total).Error; err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * limit

	if sortBy != "" && order != "" {
		if sortBy == "date" {
			sortBy = "created_at"
		}
		if sortBy == "created_at" || sortBy == "updated_at" || sortBy == "status" {
			query = query.Order(fmt.Sprintf("%s %s", sortBy, order))
		} else {
			query = query.Order("created_at desc")
		}
	} else {
		query = query.Order("created_at desc")
	}

	if err := query.Offset(offset).Limit(limit).Find(&refs).Error; err != nil {
		return nil, 0, err
	}

	if len(refs) == 0 {
		return []model.AchievementResponse{}, total, nil
	}

	var mongoIDs []primitive.ObjectID
	refMap := make(map[string]model.AchievementReference)

	for _, ref := range refs {
		oid, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		mongoIDs = append(mongoIDs, oid)
		refMap[ref.MongoAchievementID] = ref
	}

	coll := r.mongoDB.Collection("achievements")
	cursor, err := coll.Find(context.TODO(), bson.M{"_id": bson.M{"$in": mongoIDs}})
	if err != nil {
		return nil, 0, err
	}
	defer cursor.Close(context.TODO())

	var results []model.AchievementResponse
	for cursor.Next(context.TODO()) {
		var doc model.AchievementMongo
		if err := cursor.Decode(&doc); err == nil {
			hexID := doc.ID.Hex()
			if sqlRef, ok := refMap[hexID]; ok {
				results = append(results, *r.mapToResponse(sqlRef, doc))
			}
		}
	}

	return results, total, nil
}

func (r *AchievementRepo) Update(id uuid.UUID, req model.UpdateAchievementRequest) (*model.AchievementResponse, error) {
	var ref model.AchievementReference
	if err := r.pgDB.Where("status != ?", model.StatusDeleted).First(&ref, "id = ?", id).Error; err != nil {
		return nil, err
	}

	updateFields := bson.M{
		"updatedAt": time.Now(),
	}

	if req.Title != nil {
		updateFields["title"] = *req.Title
	}
	if req.AchievementType != nil {
		updateFields["achievementType"] = *req.AchievementType
	}
	if req.Description != nil {
		updateFields["description"] = *req.Description
	}
	if req.EventDate != nil {
		parsedDate, _ := time.Parse("2006-01-02", *req.EventDate)
		updateFields["details.eventDate"] = parsedDate
	}
	if req.Tags != nil {
		updateFields["tags"] = *req.Tags
	}

	if req.CompetitionDetails != nil {
		updateFields["details.competitionName"] = req.CompetitionDetails.CompetitionName
		updateFields["details.competitionLevel"] = req.CompetitionDetails.CompetitionLevel
		updateFields["details.rank"] = req.CompetitionDetails.Rank
		updateFields["details.medalType"] = req.CompetitionDetails.MedalType
	} else if req.PublicationDetails != nil {
		updateFields["details.publicationTitle"] = req.PublicationDetails.PublicationTitle
		updateFields["details.authors"] = req.PublicationDetails.Authors
		updateFields["details.publisher"] = req.PublicationDetails.Publisher
		updateFields["details.issn"] = req.PublicationDetails.ISSN
	} else if req.OrganizationDetails != nil {
		updateFields["details.organizationName"] = req.OrganizationDetails.OrganizationName
		updateFields["details.position"] = req.OrganizationDetails.Position
		if req.OrganizationDetails.StartDate != "" && req.OrganizationDetails.EndDate != "" {
			layout := "2006-01-02"
			start, err := time.Parse(layout, req.OrganizationDetails.StartDate)
			if err != nil {
				return nil, fmt.Errorf("invalid start_date format, expected YYYY-MM-DD: %w", err)
			}
			end, err := time.Parse(layout, req.OrganizationDetails.EndDate)
			if err != nil {
				return nil, fmt.Errorf("invalid end_date format, expected YYYY-MM-DD: %w", err)
			}
			updateFields["details.period"] = &model.OrganizationPeriod{
				Start: start,
				End:   end,
			}
		}
	}

	objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	coll := r.mongoDB.Collection("achievements")

	_, err := coll.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": updateFields})
	if err != nil {
		return nil, err
	}

	return r.FindByAchievementID(id)
}

func (r *AchievementRepo) UpdateStatus(id uuid.UUID, status string, verifierID *uuid.UUID, note string, points int) error {
	updates := map[string]interface{}{
		"status":     status,
		"updated_at": time.Now(),
	}

	if status == "submitted" {
		updates["submitted_at"] = time.Now()
	}
	if status == "verified" || status == "rejected" {
		updates["verified_at"] = time.Now()
		updates["verified_by"] = verifierID
		updates["rejection_note"] = note
	}

	if err := r.pgDB.Model(&model.AchievementReference{}).Where("id = ? AND status != ?", id, model.StatusDeleted).Updates(updates).Error; err != nil {
		return err
	}

	if status == "verified" {
		var ref model.AchievementReference
		if err := r.pgDB.Select("mongo_achievement_id").First(&ref, "id = ?", id).Error; err != nil {
			return err
		}

		objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		coll := r.mongoDB.Collection("achievements")

		update := bson.M{
			"$set": bson.M{"points": points},
		}

		if _, err := coll.UpdateOne(context.TODO(), bson.M{"_id": objID}, update); err != nil {
			return err
		}
	}

	return nil
}

func (r *AchievementRepo) Delete(id uuid.UUID) error {
	return r.pgDB.Model(&model.AchievementReference{}).Where("id = ?", id).Update("status", model.StatusDeleted).Error
}

func (r *AchievementRepo) AddAttachment(id uuid.UUID, attachment model.Attachment) error {
	var ref model.AchievementReference
	if err := r.pgDB.Where("status != ?", model.StatusDeleted).First(&ref, "id = ?", id).Error; err != nil {
		return err
	}

	objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	coll := r.mongoDB.Collection("achievements")

	update := bson.M{
		"$push": bson.M{"attachments": attachment},
	}
	_, err := coll.UpdateOne(context.TODO(), bson.M{"_id": objID}, update)
	return err
}

func (r *AchievementRepo) mapToResponse(ref model.AchievementReference, mongoDoc model.AchievementMongo) *model.AchievementResponse {
	var attachments []model.Attachment
	attachments = mongoDoc.Attachments

	var tags []string
	tags = mongoDoc.Tags

	studentName := ""
	if ref.Student.User.FullName != "" {
		studentName = ref.Student.User.FullName
	}

	return &model.AchievementResponse{
		ID:              ref.ID,
		MongoID:         ref.MongoAchievementID,
		StudentID:       ref.StudentID,
		StudentName:     studentName,
		Status:          string(ref.Status),
		AchievementType: mongoDoc.AchievementType,
		Title:           mongoDoc.Title,
		Description:     mongoDoc.Description,
		Details:         mongoDoc.Details,
		Attachments:     attachments,
		Tags:            tags,
		Points:          mongoDoc.Points,
		RejectionNote:   ref.RejectionNote,
		CreatedAt:       ref.CreatedAt,
		UpdatedAt:       ref.UpdatedAt,
	}
}

func (r *AchievementRepo) GetOwnerID(achievementID uuid.UUID) (uuid.UUID, error) {
	var ref model.AchievementReference
	err := r.pgDB.Preload("Student").Where("status != ?", model.StatusDeleted).First(&ref, "id = ?", achievementID).Error
	if err != nil {
		return uuid.Nil, err
	}
	return ref.Student.UserID, nil
}

func (r *AchievementRepo) IsAdvisor(lecturerUserID uuid.UUID, achievementID uuid.UUID) (bool, error) {
	var ref model.AchievementReference
	err := r.pgDB.Joins("JOIN students ON students.id = achievement_references.student_id").
		Joins("JOIN lecturers ON lecturers.id = students.advisor_id").
		Where("achievement_references.id = ? AND lecturers.user_id = ? AND achievement_references.status != ?", achievementID, lecturerUserID, model.StatusDeleted).
		First(&ref).Error

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, err
	}

	return true, nil
}

func (r *AchievementRepo) GetStatus(id uuid.UUID) (string, error) {
	var status string
	err := r.pgDB.Model(&model.AchievementReference{}).
		Where("id = ? AND status != ?", id, model.StatusDeleted).
		Select("status").
		Scan(&status).Error
	if err != nil {
		return "", err
	}
	return status, nil
}

func (r *AchievementRepo) GetHistory(id uuid.UUID) (*model.AchievementHistoryResponse, error) {
	var ref model.AchievementReference
	if err := r.pgDB.Preload("Verifier").Where("status != ?", model.StatusDeleted).First(&ref, "id = ?", id).Error; err != nil {
		return nil, err
	}

	var mongoDetail model.AchievementMongo
	objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	coll := r.mongoDB.Collection("achievements")
	err := coll.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&mongoDetail)
	if err != nil {
		return nil, errors.New("detail data missing in NoSQL")
	}

	verifierName := ""
	if ref.Verifier != nil {
		verifierName = ref.Verifier.FullName
	}

	return &model.AchievementHistoryResponse{
		ID:            ref.ID,
		Title:         mongoDetail.Title,
		Status:        string(ref.Status),
		CreatedAt:     ref.CreatedAt,
		SubmittedAt:   ref.SubmittedAt,
		VerifiedAt:    ref.VerifiedAt,
		VerifierName:  verifierName,
		RejectionNote: ref.RejectionNote,
		Points:        mongoDetail.Points,
	}, nil
}
