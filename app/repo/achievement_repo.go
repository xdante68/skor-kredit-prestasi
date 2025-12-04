package repo

import (
	"context"
	"errors"
	"fiber/skp/app/model"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"gorm.io/gorm"
)

type AchievementRepo struct {
	pgDB    *gorm.DB
	mongoDB *mongo.Database
}

func NewAchievementRepo(pgDB *gorm.DB, mongoDB *mongo.Database) *AchievementRepo {
	return &AchievementRepo{pgDB: pgDB, mongoDB: mongoDB}
}

func (r *AchievementRepo) Create(studentID uuid.UUID, req model.CreateAchievementRequest) (*model.AchievementResponse, error) {
	var details interface{}
	var detailsMap map[string]interface{}
	if req.CompetitionDetails != nil {
		details = req.CompetitionDetails
		detailsMap = map[string]interface{}{
			"competition_name":  req.CompetitionDetails.CompetitionName,
			"competition_level": req.CompetitionDetails.CompetitionLevel,
			"rank":              req.CompetitionDetails.Rank,
			"medal_type":        req.CompetitionDetails.MedalType,
		}
	} else if req.PublicationDetails != nil {
		details = req.PublicationDetails
		detailsMap = map[string]interface{}{
			"publication_title": req.PublicationDetails.PublicationTitle,
			"authors":           req.PublicationDetails.Authors,
			"publisher":         req.PublicationDetails.Publisher,
			"issn":              req.PublicationDetails.ISSN,
		}
	} else if req.OrganizationDetails != nil {
		details = req.OrganizationDetails
		detailsMap = map[string]interface{}{
			"organization_name": req.OrganizationDetails.OrganizationName,
			"position":          req.OrganizationDetails.Position,
			"start_date":        req.OrganizationDetails.StartDate,
			"end_date":          req.OrganizationDetails.EndDate,
		}
	} else {
		details = bson.M{}
		detailsMap = map[string]interface{}{}
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
		Details:         detailsMap,
		Tags:            req.Tags,
		Attachments:     []model.Attachment{},
		Points:          0,
		CreatedAt:       now,
		UpdatedAt:       now,
	}, nil
}

func (r *AchievementRepo) FindByID(id uuid.UUID) (*model.AchievementResponse, error) {
	var ref model.AchievementReference
	if err := r.pgDB.Preload("Student.User").Where("status != ?", model.StatusDeleted).First(&ref, "id = ?", id).Error; err != nil {
		return nil, err
	}

	var mongoDetail bson.M
	objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)

	coll := r.mongoDB.Collection("achievements")
	err := coll.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&mongoDetail)

	if err != nil {
		return nil, errors.New("detail data missing in NoSQL")
	}

	return r.mapToResponse(ref, mongoDetail), nil
}

func (r *AchievementRepo) FindAll(role string, userID uuid.UUID) ([]model.AchievementResponse, error) {
	var refs []model.AchievementReference

	query := r.pgDB.Preload("Student.User").Where("status != ?", model.StatusDeleted)

	if role == "Mahasiswa" {
		var student model.Student
		if err := r.pgDB.Where("user_id = ?", userID).First(&student).Error; err != nil {
			return nil, err
		}
		query = query.Where("student_id = ?", student.ID)

	} else if role == "Dosen Wali" {
		var lecturer model.Lecturer
		if err := r.pgDB.Where("user_id = ?", userID).First(&lecturer).Error; err != nil {
			return nil, err
		}
		query = query.Joins("JOIN students ON students.id = achievement_references.student_id").
			Where("students.advisor_id = ?", lecturer.ID).
			Where("achievement_references.status != ?", model.StatusDraft)
	}
	if role == "Admin" {
		query = query.Where("achievement_references.status != ?", model.StatusDraft)
	}

	if err := query.Find(&refs).Error; err != nil {
		return nil, err
	}

	if len(refs) == 0 {
		return []model.AchievementResponse{}, nil
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
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []model.AchievementResponse
	for cursor.Next(context.TODO()) {
		var doc bson.M
		if err := cursor.Decode(&doc); err == nil {
			hexID := doc["_id"].(primitive.ObjectID).Hex()
			if sqlRef, ok := refMap[hexID]; ok {
				results = append(results, *r.mapToResponse(sqlRef, doc))
			}
		}
	}

	return results, nil
}

func (r *AchievementRepo) Update(id uuid.UUID, req model.UpdateAchievementRequest) error {
	var ref model.AchievementReference
	if err := r.pgDB.Where("status != ?", model.StatusDeleted).First(&ref, "id = ?", id).Error; err != nil {
		return err
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
	if req.Tags != nil {
		updateFields["tags"] = *req.Tags
	}

	if req.CompetitionDetails != nil {
		updateFields["details"] = req.CompetitionDetails
	} else if req.PublicationDetails != nil {
		updateFields["details"] = req.PublicationDetails
	} else if req.OrganizationDetails != nil {
		updateFields["details"] = req.OrganizationDetails
	}

	objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
	coll := r.mongoDB.Collection("achievements")

	updateData := bson.M{
		"$set": updateFields,
	}

	_, err := coll.UpdateOne(context.TODO(), bson.M{"_id": objID}, updateData)
	return err
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

func (r *AchievementRepo) mapToResponse(ref model.AchievementReference, mongoDoc bson.M) *model.AchievementResponse {
	var attachments []model.Attachment
	if rawArr, ok := mongoDoc["attachments"].(primitive.A); ok {
		for _, item := range rawArr {
			if m, ok := item.(primitive.M); ok {
				attachments = append(attachments, model.Attachment{
					FileName: m["fileName"].(string),
					FileURL:  m["fileUrl"].(string),
					FileType: m["fileType"].(string),
				})
			}
		}
	}

	var tags []string
	if rawTags, ok := mongoDoc["tags"].(primitive.A); ok {
		for _, t := range rawTags {
			tags = append(tags, t.(string))
		}
	}

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
		AchievementType: mongoDoc["achievementType"].(string),
		Title:           mongoDoc["title"].(string),
		Description:     mongoDoc["description"].(string),
		Details:         mongoDoc["details"].(primitive.M),
		Attachments:     attachments,
		Tags:            tags,
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

func (r *AchievementRepo) IsAdvisor(lecturerUserID uuid.UUID, achievementID uuid.UUID) bool {
	var ref model.AchievementReference
	err := r.pgDB.Joins("JOIN students ON students.id = achievement_references.student_id").
		Joins("JOIN lecturers ON lecturers.id = students.advisor_id").
		Where("achievement_references.id = ? AND lecturers.user_id = ? AND achievement_references.status != ?", achievementID, lecturerUserID, model.StatusDeleted).
		First(&ref).Error

	return err == nil
}
