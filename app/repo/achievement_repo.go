package repo

import (
	"context"
	"database/sql"
	"errors"
	"fiber/skp/app/model"
	"fmt"
	"time"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
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
	pgDB    *sql.DB
	mongoDB *mongo.Database
}

func NewAchievementRepo(pgDB *sql.DB, mongoDB *mongo.Database) *AchievementRepo {
	return &AchievementRepo{pgDB: pgDB, mongoDB: mongoDB}
}

var achievementSortWhitelist = map[string]string{
	"created_at": "ar.created_at",
	"updated_at": "ar.updated_at",
	"status":     "ar.status",
	"date":       "ar.created_at",
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
		"attachments":     []interface{}{},
		"points":          0,
		"createdAt":       now,
		"updatedAt":       now,
	}

	coll := r.mongoDB.Collection("achievements")
	res, err := coll.InsertOne(context.TODO(), mongoData)
	if err != nil {
		return nil, err
	}
	oid := res.InsertedID.(primitive.ObjectID)

	query := `
		INSERT INTO achievement_references (student_id, mongo_achievement_id, status, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5)
		RETURNING id`

	var pgID uuid.UUID
	err = r.pgDB.QueryRow(query, studentID, oid.Hex(), "draft", now, now).Scan(&pgID)
	if err != nil {
		coll.DeleteOne(context.TODO(), bson.M{"_id": oid})
		return nil, err
	}

	return &model.AchievementResponse{
		ID:              pgID,
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
	query := `
		SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status, ar.rejection_note, ar.created_at, ar.updated_at,
		       s.id, u.full_name
		FROM achievement_references ar
		JOIN students s ON s.id = ar.student_id
		JOIN users u ON u.id = s.user_id
		WHERE ar.id = $1 AND ar.status != $2`

	var ref model.AchievementReference
	var studentID uuid.UUID
	var studentFullName sql.NullString
	var rejectionNote sql.NullString

	err := r.pgDB.QueryRow(query, id, model.StatusDeleted).Scan(
		&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status, &rejectionNote, &ref.CreatedAt, &ref.UpdatedAt,
		&studentID, &studentFullName,
	)
	if err != nil {
		return nil, err
	}

	if rejectionNote.Valid {
		ref.RejectionNote = rejectionNote.String
	}

	ref.Student.ID = studentID
	if studentFullName.Valid {
		ref.Student.User.FullName = studentFullName.String
	}

	var mongoDetail model.AchievementMongo
	objID, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)

	coll := r.mongoDB.Collection("achievements")
	err = coll.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&mongoDetail)

	if err != nil {
		return nil, errors.New("detail data missing in NoSQL")
	}

	return r.mapToResponse(ref, mongoDetail), nil
}

func (r *AchievementRepo) FindAll(role string, userID uuid.UUID, page, limit int, search, sortBy, order string) ([]model.AchievementResponse, int64, error) {
	var total int64
	var mongoIDs []string
	if search != "" {
		coll := r.mongoDB.Collection("achievements")
		filter := bson.M{"title": bson.M{"$regex": search, "$options": "i"}}
		cursor, err := coll.Find(context.TODO(), filter)
		if err != nil {
			return nil, 0, err
		}
		defer cursor.Close(context.TODO())

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
	}

	args := []interface{}{}
	argIndex := 1

	countQuery := `SELECT COUNT(*) FROM achievement_references ar WHERE ar.status != $1`
	args = append(args, model.StatusDeleted)
	argIndex++

	if search != "" && len(mongoIDs) > 0 {
		countQuery += fmt.Sprintf(" AND ar.mongo_achievement_id = ANY($%d)", argIndex)
		args = append(args, mongoIDs)
		argIndex++
	}

	if role == model.RoleMahasiswa {
		countQuery += fmt.Sprintf(" AND ar.student_id = (SELECT id FROM students WHERE user_id = $%d)", argIndex)
		args = append(args, userID)
		argIndex++
	} else if role == model.RoleDosenWali {
		countQuery += fmt.Sprintf(" AND ar.student_id IN (SELECT s.id FROM students s JOIN lecturers l ON s.advisor_id = l.id WHERE l.user_id = $%d) AND ar.status != $%d", argIndex, argIndex+1)
		args = append(args, userID, model.StatusDraft)
		argIndex += 2
	} else if role == model.RoleAdmin {
		countQuery += fmt.Sprintf(" AND ar.status != $%d", argIndex)
		args = append(args, model.StatusDraft)
		argIndex++
	}

	err := r.pgDB.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, err
	}

	mainQuery := `
		SELECT ar.id, ar.student_id, ar.mongo_achievement_id, ar.status, ar.rejection_note, ar.created_at, ar.updated_at,
		       s.id, u.full_name
		FROM achievement_references ar
		JOIN students s ON s.id = ar.student_id
		JOIN users u ON u.id = s.user_id
		WHERE ar.status != $1`

	selectArgs := []interface{}{model.StatusDeleted}
	selectArgIndex := 2

	if search != "" && len(mongoIDs) > 0 {
		mainQuery += fmt.Sprintf(" AND ar.mongo_achievement_id = ANY($%d)", selectArgIndex)
		selectArgs = append(selectArgs, mongoIDs)
		selectArgIndex++
	}

	if role == model.RoleMahasiswa {
		mainQuery += fmt.Sprintf(" AND ar.student_id = (SELECT id FROM students WHERE user_id = $%d)", selectArgIndex)
		selectArgs = append(selectArgs, userID)
		selectArgIndex++
	} else if role == model.RoleDosenWali {
		mainQuery += fmt.Sprintf(" AND ar.student_id IN (SELECT s.id FROM students s JOIN lecturers l ON s.advisor_id = l.id WHERE l.user_id = $%d) AND ar.status != $%d", selectArgIndex, selectArgIndex+1)
		selectArgs = append(selectArgs, userID, model.StatusDraft)
		selectArgIndex += 2
	} else if role == model.RoleAdmin {
		mainQuery += fmt.Sprintf(" AND ar.status != $%d", selectArgIndex)
		selectArgs = append(selectArgs, model.StatusDraft)
		selectArgIndex++
	}

	if order != "asc" && order != "desc" {
		order = "desc"
	}
	if sortColumn, ok := achievementSortWhitelist[sortBy]; ok {
		mainQuery += fmt.Sprintf(" ORDER BY %s %s", sortColumn, order)
	} else {
		mainQuery += " ORDER BY ar.created_at DESC"
	}

	offset := (page - 1) * limit
	mainQuery += fmt.Sprintf(" LIMIT $%d OFFSET $%d", selectArgIndex, selectArgIndex+1)
	selectArgs = append(selectArgs, limit, offset)

	rows, err := r.pgDB.Query(mainQuery, selectArgs...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	var refs []model.AchievementReference
	refMap := make(map[string]model.AchievementReference)

	for rows.Next() {
		var ref model.AchievementReference
		var studentID uuid.UUID
		var studentFullName, rejectionNote sql.NullString

		if err := rows.Scan(
			&ref.ID, &ref.StudentID, &ref.MongoAchievementID, &ref.Status, &rejectionNote, &ref.CreatedAt, &ref.UpdatedAt,
			&studentID, &studentFullName,
		); err != nil {
			return nil, 0, err
		}

		if rejectionNote.Valid {
			ref.RejectionNote = rejectionNote.String
		}

		ref.Student.ID = studentID
		if studentFullName.Valid {
			ref.Student.User.FullName = studentFullName.String
		}

		refs = append(refs, ref)
		refMap[ref.MongoAchievementID] = ref
	}

	if len(refs) == 0 {
		return []model.AchievementResponse{}, total, nil
	}

	var mongoOIDs []primitive.ObjectID
	for _, ref := range refs {
		oid, _ := primitive.ObjectIDFromHex(ref.MongoAchievementID)
		mongoOIDs = append(mongoOIDs, oid)
	}

	coll := r.mongoDB.Collection("achievements")
	cursor, err := coll.Find(context.TODO(), bson.M{"_id": bson.M{"$in": mongoOIDs}})
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
	query := `SELECT mongo_achievement_id FROM achievement_references WHERE id = $1 AND status != $2`
	var mongoID string
	err := r.pgDB.QueryRow(query, id, model.StatusDeleted).Scan(&mongoID)
	if err != nil {
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

	objID, _ := primitive.ObjectIDFromHex(mongoID)
	coll := r.mongoDB.Collection("achievements")

	_, err = coll.UpdateOne(context.TODO(), bson.M{"_id": objID}, bson.M{"$set": updateFields})
	if err != nil {
		return nil, err
	}

	_, err = r.pgDB.Exec("UPDATE achievement_references SET updated_at = $1 WHERE id = $2", time.Now(), id)
	if err != nil {
		return nil, err
	}

	return r.FindByAchievementID(id)
}

func (r *AchievementRepo) UpdateStatus(id uuid.UUID, status string, verifierID *uuid.UUID, note string, points int) error {
	now := time.Now()

	var query string
	var args []interface{}

	if status == "submitted" {
		query = `UPDATE achievement_references SET status = $1, submitted_at = $2, updated_at = $3 WHERE id = $4 AND status != $5`
		args = []interface{}{status, now, now, id, model.StatusDeleted}
	} else if status == "verified" || status == "rejected" {
		query = `UPDATE achievement_references SET status = $1, verified_at = $2, verified_by = $3, rejection_note = $4, updated_at = $5 WHERE id = $6 AND status != $7`
		args = []interface{}{status, now, verifierID, note, now, id, model.StatusDeleted}
	} else {
		query = `UPDATE achievement_references SET status = $1, updated_at = $2 WHERE id = $3 AND status != $4`
		args = []interface{}{status, now, id, model.StatusDeleted}
	}

	_, err := r.pgDB.Exec(query, args...)
	if err != nil {
		return err
	}

	if status == "verified" {
		var mongoID string
		err := r.pgDB.QueryRow("SELECT mongo_achievement_id FROM achievement_references WHERE id = $1", id).Scan(&mongoID)
		if err != nil {
			return err
		}

		objID, _ := primitive.ObjectIDFromHex(mongoID)
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
	query := `UPDATE achievement_references SET status = $1 WHERE id = $2`
	_, err := r.pgDB.Exec(query, model.StatusDeleted, id)
	return err
}

func (r *AchievementRepo) AddAttachment(id uuid.UUID, attachment model.Attachment) error {
	var mongoID string
	err := r.pgDB.QueryRow("SELECT mongo_achievement_id FROM achievement_references WHERE id = $1 AND status != $2", id, model.StatusDeleted).Scan(&mongoID)
	if err != nil {
		return err
	}

	objID, _ := primitive.ObjectIDFromHex(mongoID)
	coll := r.mongoDB.Collection("achievements")

	update := bson.M{
		"$push": bson.M{"attachments": attachment},
	}
	_, err = coll.UpdateOne(context.TODO(), bson.M{"_id": objID}, update)
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
	query := `
		SELECT s.user_id
		FROM achievement_references ar
		JOIN students s ON s.id = ar.student_id
		WHERE ar.id = $1 AND ar.status != $2`

	var ownerID uuid.UUID
	err := r.pgDB.QueryRow(query, achievementID, model.StatusDeleted).Scan(&ownerID)
	if err != nil {
		return uuid.Nil, err
	}
	return ownerID, nil
}

func (r *AchievementRepo) IsAdvisor(lecturerUserID uuid.UUID, achievementID uuid.UUID) (bool, error) {
	query := `
		SELECT COUNT(*)
		FROM achievement_references ar
		JOIN students s ON s.id = ar.student_id
		JOIN lecturers l ON l.id = s.advisor_id
		WHERE ar.id = $1 AND l.user_id = $2 AND ar.status != $3`

	var count int64
	err := r.pgDB.QueryRow(query, achievementID, lecturerUserID, model.StatusDeleted).Scan(&count)
	if err != nil {
		return false, err
	}
	return count > 0, nil
}

func (r *AchievementRepo) GetStatus(id uuid.UUID) (string, error) {
	var status string
	query := `SELECT status FROM achievement_references WHERE id = $1 AND status != $2`
	err := r.pgDB.QueryRow(query, id, model.StatusDeleted).Scan(&status)
	if err != nil {
		return "", err
	}
	return status, nil
}

func (r *AchievementRepo) GetHistory(id uuid.UUID) (*model.AchievementHistoryResponse, error) {
	query := `
		SELECT ar.id, ar.mongo_achievement_id, ar.status, ar.created_at, ar.submitted_at, ar.verified_at, ar.rejection_note,
		       u.full_name
		FROM achievement_references ar
		LEFT JOIN users u ON u.id = ar.verified_by
		WHERE ar.id = $1 AND ar.status != $2`

	var mongoID string
	var status string
	var createdAt time.Time
	var submittedAt, verifiedAt sql.NullTime
	var rejectionNote, verifierName sql.NullString
	var pgID uuid.UUID

	err := r.pgDB.QueryRow(query, id, model.StatusDeleted).Scan(
		&pgID, &mongoID, &status, &createdAt, &submittedAt, &verifiedAt, &rejectionNote, &verifierName,
	)
	if err != nil {
		return nil, err
	}

	var mongoDetail model.AchievementMongo
	objID, _ := primitive.ObjectIDFromHex(mongoID)
	coll := r.mongoDB.Collection("achievements")
	err = coll.FindOne(context.TODO(), bson.M{"_id": objID}).Decode(&mongoDetail)
	if err != nil {
		return nil, errors.New("detail data missing in NoSQL")
	}

	verName := ""
	if verifierName.Valid {
		verName = verifierName.String
	}

	rejNote := ""
	if rejectionNote.Valid {
		rejNote = rejectionNote.String
	}

	resp := &model.AchievementHistoryResponse{
		ID:            pgID,
		Title:         mongoDetail.Title,
		Status:        status,
		CreatedAt:     createdAt,
		VerifierName:  verName,
		RejectionNote: rejNote,
		Points:        mongoDetail.Points,
	}

	if submittedAt.Valid {
		resp.SubmittedAt = &submittedAt.Time
	}
	if verifiedAt.Valid {
		resp.VerifiedAt = &verifiedAt.Time
	}

	return resp, nil
}
