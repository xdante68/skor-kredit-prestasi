package repo

import (
	"context"
	"database/sql"
	"fiber/skp/app/model"

	"github.com/google/uuid"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type ReportRepository interface {
	GetMongoIDs(role string, userID uuid.UUID) ([]primitive.ObjectID, error)
	GetStatsByPeriod(role string, userID uuid.UUID) ([]model.StatItem, error)
	GetTopStudents() ([]model.TopStudent, error)
	GetStudentProfile(studentID uuid.UUID) (*model.TopStudent, error)
	GetMongoIDsByStudentID(studentID uuid.UUID) ([]primitive.ObjectID, error)
	GetStatsByType(ids []primitive.ObjectID) ([]model.StatItem, error)
	GetStatsByLevel(ids []primitive.ObjectID) ([]model.StatItem, error)
}

type ReportRepo struct {
	pgDB    *sql.DB
	mongoDB *mongo.Database
}

func NewReportRepo(pgDB *sql.DB, mongoDB *mongo.Database) *ReportRepo {
	return &ReportRepo{pgDB: pgDB, mongoDB: mongoDB}
}

func (r *ReportRepo) GetMongoIDs(role string, userID uuid.UUID) ([]primitive.ObjectID, error) {
	var query string
	var args []interface{}

	switch role {
	case model.RoleAdmin:
		query = `
			SELECT ar.mongo_achievement_id 
			FROM achievement_references ar
			WHERE ar.status = 'verified'`
	case model.RoleDosenWali:
		query = `
			SELECT ar.mongo_achievement_id 
			FROM achievement_references ar
			JOIN students s ON ar.student_id = s.id
			JOIN lecturers l ON s.advisor_id = l.id
			WHERE ar.status = 'verified' AND l.user_id = $1`
		args = append(args, userID)
	default:
		query = `
			SELECT ar.mongo_achievement_id 
			FROM achievement_references ar
			JOIN students s ON ar.student_id = s.id
			WHERE ar.status = 'verified' AND s.user_id = $1`
		args = append(args, userID)
	}

	rows, err := r.pgDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var objectIDs []primitive.ObjectID
	for rows.Next() {
		var mongoID string
		if err := rows.Scan(&mongoID); err != nil {
			continue
		}
		if oid, err := primitive.ObjectIDFromHex(mongoID); err == nil {
			objectIDs = append(objectIDs, oid)
		}
	}
	return objectIDs, nil
}

func (r *ReportRepo) GetMongoIDsByStudentID(studentID uuid.UUID) ([]primitive.ObjectID, error) {
	query := `
		SELECT ar.mongo_achievement_id 
		FROM achievement_references ar
		WHERE ar.student_id = $1 AND ar.status = 'verified'`

	rows, err := r.pgDB.Query(query, studentID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var objectIDs []primitive.ObjectID
	for rows.Next() {
		var mongoID string
		if err := rows.Scan(&mongoID); err != nil {
			continue
		}
		if oid, err := primitive.ObjectIDFromHex(mongoID); err == nil {
			objectIDs = append(objectIDs, oid)
		}
	}
	return objectIDs, nil
}

func (r *ReportRepo) GetStatsByPeriod(role string, userID uuid.UUID) ([]model.StatItem, error) {
	var query string
	var args []interface{}

	baseQuery := `
		SELECT TO_CHAR(ar.created_at, 'YYYY') as label, COUNT(ar.id) as count
		FROM achievement_references ar`

	groupOrder := `
		GROUP BY TO_CHAR(ar.created_at, 'YYYY')
		ORDER BY label DESC`

	switch role {
	case model.RoleAdmin:
		query = baseQuery + ` WHERE ar.status = 'verified' ` + groupOrder
	case model.RoleDosenWali:
		query = baseQuery + `
			JOIN students s ON ar.student_id = s.id
			JOIN lecturers l ON s.advisor_id = l.id
			WHERE ar.status = 'verified' AND l.user_id = $1 ` + groupOrder
		args = append(args, userID)
	default:
		query = baseQuery + `
			JOIN students s ON ar.student_id = s.id
			WHERE ar.status = 'verified' AND s.user_id = $1 ` + groupOrder
		args = append(args, userID)
	}

	rows, err := r.pgDB.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.StatItem
	for rows.Next() {
		var item model.StatItem
		if err := rows.Scan(&item.Label, &item.Count); err != nil {
			continue
		}
		results = append(results, item)
	}
	return results, nil
}

func (r *ReportRepo) GetTopStudents() ([]model.TopStudent, error) {
	query := `
		SELECT 
			s.student_id,
			u.full_name as student_name, 
			s.program_study as program, 
			COUNT(ar.id) as total_achievements,
			COALESCE(SUM(ar.points), 0) as total_points
		FROM achievement_references ar
		JOIN students s ON ar.student_id = s.id
		JOIN users u ON s.user_id = u.id
		WHERE ar.status = 'verified'
		GROUP BY s.student_id, u.full_name, s.program_study
		ORDER BY total_achievements DESC, total_points DESC
		LIMIT 5`

	rows, err := r.pgDB.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []model.TopStudent
	for rows.Next() {
		var student model.TopStudent
		if err := rows.Scan(
			&student.StudentID,
			&student.StudentName,
			&student.Program,
			&student.TotalAchievements,
			&student.TotalPoints,
		); err != nil {
			continue
		}
		results = append(results, student)
	}
	return results, nil
}

func (r *ReportRepo) GetStudentProfile(studentID uuid.UUID) (*model.TopStudent, error) {
	query := `
		SELECT 
			s.student_id,
			u.full_name as student_name, 
			s.program_study as program,
			COUNT(ar.id) as total_achievements,
			COALESCE(SUM(ar.points), 0) as total_points
		FROM students s
		JOIN users u ON s.user_id = u.id
		LEFT JOIN achievement_references ar ON ar.student_id = s.id AND ar.status = 'verified'
		WHERE s.id = $1
		GROUP BY s.student_id, u.full_name, s.program_study`

	var profile model.TopStudent
	err := r.pgDB.QueryRow(query, studentID).Scan(
		&profile.StudentID,
		&profile.StudentName,
		&profile.Program,
		&profile.TotalAchievements,
		&profile.TotalPoints,
	)
	if err != nil {
		return nil, err
	}
	return &profile, nil
}

func (r *ReportRepo) GetStatsByType(ids []primitive.ObjectID) ([]model.StatItem, error) {
	if len(ids) == 0 {
		return []model.StatItem{}, nil
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}}}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$achievementType"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
	}
	return r.aggregateMongo(pipeline)
}

func (r *ReportRepo) GetStatsByLevel(ids []primitive.ObjectID) ([]model.StatItem, error) {
	if len(ids) == 0 {
		return []model.StatItem{}, nil
	}

	pipeline := mongo.Pipeline{
		{{Key: "$match", Value: bson.D{
			{Key: "_id", Value: bson.D{{Key: "$in", Value: ids}}},
			{Key: "achievementType", Value: "competition"},
		}}},
		{{Key: "$group", Value: bson.D{
			{Key: "_id", Value: "$details.competitionLevel"},
			{Key: "count", Value: bson.D{{Key: "$sum", Value: 1}}},
		}}},
		{{Key: "$sort", Value: bson.D{{Key: "count", Value: -1}}}},
	}
	return r.aggregateMongo(pipeline)
}

func (r *ReportRepo) aggregateMongo(pipeline mongo.Pipeline) ([]model.StatItem, error) {
	coll := r.mongoDB.Collection("achievements")
	cursor, err := coll.Aggregate(context.TODO(), pipeline)
	if err != nil {
		return nil, err
	}
	defer cursor.Close(context.TODO())

	var results []model.StatItem
	if err = cursor.All(context.TODO(), &results); err != nil {
		return nil, err
	}
	return results, nil
}
