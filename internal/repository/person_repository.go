package repository

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sirupsen/logrus"
	"person-service/internal/domain"
	"strings"
)

var (
	ErrNotFound = errors.New("not found")
)

type PersonRepositoryInterface interface {
	Create(ctx context.Context, person *domain.Person) (int64, error)
	GetById(ctx context.Context, id int64) (*domain.Person, error)
	GetAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*domain.Person, int, error)
	Update(ctx context.Context, id int64, person *domain.Person) error
	Delete(ctx context.Context, id int64) error
}

type PersonRepository struct {
	db  *pgxpool.Pool
	log *logrus.Logger
}

func NewPersonRepository(db *pgxpool.Pool, log *logrus.Logger) PersonRepositoryInterface {
	return &PersonRepository{
		db:  db,
		log: log,
	}
}

func (r *PersonRepository) Create(ctx context.Context, person *domain.Person) (int64, error) {
	query := `
		INSERT INTO people (name, surname, patronymic, age, gender, nationality)
		VALUES ($1, $2, $3, $4, $5, $6)
		RETURNING id
    `
	var id int64
	err := r.db.QueryRow(ctx, query,
		person.Name,
		person.Surname,
		person.Patronymic,
		person.Age,
		person.Gender,
		person.Nationality,
	).Scan(&id)
	if err != nil {
		logrus.Errorf("Failed to create person with name %s: %v", person.Name, err)
		return 0, err
	}
	logrus.Debugf("Created person with ID: %d", id)
	return id, nil
}

func (r *PersonRepository) GetById(ctx context.Context, id int64) (*domain.Person, error) {
	query := "SELECT id, name, surname, patronymic, age, gender, nationality, created_at FROM people WHERE id = $1"
	person := &domain.Person{}
	err := r.db.QueryRow(ctx, query, id).Scan(
		&person.ID, &person.Name, &person.Surname, &person.Patronymic, &person.Age, &person.Gender, &person.Nationality, &person.CreatedAt,
	)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, ErrNotFound
		}
		logrus.WithError(err).Errorf("Failed to get person by ID: %d", id)
		return nil, fmt.Errorf("failed to get person: %w", err)
	}
	logrus.Debugf("Retrieved person with ID %d: %+v", id, person)
	return person, nil
}

func (r *PersonRepository) GetAll(ctx context.Context, filters map[string]interface{}, limit, offset int) ([]*domain.Person, int, error) {
	// Формируем запрос для получения записей
	query := "SELECT id, name, surname, patronymic, age, gender, nationality, created_at FROM people"
	var args []interface{}
	var conditions []string
	argIndex := 1

	if name, ok := filters["name"]; ok {
		conditions = append(conditions, fmt.Sprintf("name ILIKE $%d", argIndex))
		args = append(args, "%"+name.(string)+"%")
		argIndex++
	}
	if surname, ok := filters["surname"]; ok {
		conditions = append(conditions, fmt.Sprintf("surname ILIKE $%d", argIndex))
		args = append(args, "%"+surname.(string)+"%")
		argIndex++
	}
	if age, ok := filters["age"]; ok {
		conditions = append(conditions, fmt.Sprintf("age = $%d", argIndex))
		args = append(args, age.(int))
		argIndex++
	}
	if gender, ok := filters["gender"]; ok {
		conditions = append(conditions, fmt.Sprintf("gender = $%d", argIndex))
		args = append(args, gender.(string))
		argIndex++
	}
	if nationality, ok := filters["nationality"]; ok {
		conditions = append(conditions, fmt.Sprintf("nationality ILIKE $%d", argIndex))
		args = append(args, "%"+nationality.(string)+"%")
		argIndex++
	}

	// Сохраняем условия для COUNT-запроса
	countQuery := "SELECT COUNT(*) FROM people"
	countArgs := args
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
		countQuery += " WHERE " + strings.Join(conditions, " AND ")
	}

	// Добавляем LIMIT и OFFSET для основного запроса
	query += fmt.Sprintf(" ORDER BY created_at DESC LIMIT $%d OFFSET $%d", argIndex, argIndex+1)
	args = append(args, limit, offset)

	// Выполняем COUNT-запрос
	var total int
	err := r.db.QueryRow(ctx, countQuery, countArgs...).Scan(&total)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"query": countQuery,
			"args":  countArgs,
			"error": err,
		}).Error("Failed to count people")
		return nil, 0, fmt.Errorf("failed to count people: %w", err)
	}

	// Выполняем основной запрос
	rows, err := r.db.Query(ctx, query, args...)
	if err != nil {
		r.log.WithFields(logrus.Fields{
			"query": query,
			"args":  args,
			"error": err,
		}).Error("Failed to get people")
		return nil, 0, fmt.Errorf("failed to get people: %w", err)
	}
	defer rows.Close()

	var people []*domain.Person
	for rows.Next() {
		person := &domain.Person{}
		var patronymic sql.NullString
		if err := rows.Scan(
			&person.ID,
			&person.Name,
			&person.Surname,
			&patronymic,
			&person.Age,
			&person.Gender,
			&person.Nationality,
			&person.CreatedAt,
		); err != nil {
			r.log.WithError(err).Error("Failed to scan person")
			return nil, 0, fmt.Errorf("failed to scan person: %w", err)
		}
		if patronymic.Valid {
			person.Patronymic = &patronymic.String
		}
		people = append(people, person)
	}

	if err := rows.Err(); err != nil {
		r.log.WithError(err).Error("Rows error")
		return nil, 0, fmt.Errorf("rows error: %w", err)
	}

	r.log.WithField("count", len(people)).Debug("Retrieved people")
	return people, total, nil
}

func (r *PersonRepository) Update(ctx context.Context, id int64, person *domain.Person) error {
	query := `
        UPDATE people
        SET name = $1, surname = $2, patronymic = $3, age = $4, gender = $5, nationality = $6
        WHERE id = $7
    `
	result, err := r.db.Exec(ctx, query,
		person.Name, person.Surname, person.Patronymic, person.Age, person.Gender, person.Nationality, id,
	)

	if err != nil {
		logrus.Errorf("Failed to update person ID %d: %v", id, err)
		return err
	}

	if result.RowsAffected() == 0 {
		logrus.Warnf("No person updated with ID %d", id)
		return ErrNotFound
	}
	logrus.Debugf("Updated person with ID: %d", id)
	return nil
}

func (r *PersonRepository) Delete(ctx context.Context, id int64) error {
	query := "DELETE FROM people WHERE id = $1"
	result, err := r.db.Exec(ctx, query, id)

	if err != nil {
		logrus.Errorf("Failed to delete person ID %d: %v", id, err)
		return err
	}
	if result.RowsAffected() == 0 {
		logrus.Warnf("No person deleted with ID %d", id)
		return ErrNotFound
	}
	logrus.Debugf("Deleted person with ID: %d", id)
	return nil
}
