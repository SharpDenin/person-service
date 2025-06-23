package service

import (
	"context"
	"errors"
	"fmt"
	"github.com/sirupsen/logrus"
	"os"
	"person-service/internal/domain"
	"person-service/internal/repository"
	"sync"
)

type PersonServiceInterface interface {
	Create(ctx context.Context, person *domain.Person) (int64, error)
	GetById(ctx context.Context, id int64) (*domain.Person, error)
	GetAll(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.Person, int, error)
	Update(ctx context.Context, id int64, person *domain.Person) error
	Delete(ctx context.Context, id int64) error
}
type PersonService struct {
	repo   repository.PersonRepositoryInterface
	client *EnrichmentClient
	log    *logrus.Logger
}

func NewPersonService(repo repository.PersonRepositoryInterface, log *logrus.Logger) PersonServiceInterface {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &PersonService{
		repo:   repo,
		client: NewEnrichmentClient(log),
		log:    log,
	}
}

func (s *PersonService) Create(ctx context.Context, person *domain.Person) (int64, error) {
	if person.Name == "" || person.Surname == "" {
		s.log.Error("Name and surname are required")
		return 0, fmt.Errorf("name and surname are required")
	}
	s.log.Debugf("Creating person with name %s and surname %s", person.Name, person.Surname)

	var wg sync.WaitGroup
	wg.Add(3)

	go func() {
		defer wg.Done()
		if age, err := s.client.GetAge(ctx, person.Name); err == nil {
			person.Age = age
		} else {
			s.log.WithError(err).Warn("Age enrichment failed")
		}
	}()

	go func() {
		defer wg.Done()
		if gender, err := s.client.GetGender(ctx, person.Name); err == nil {
			person.Gender = gender
		} else {
			s.log.WithError(err).Warn("Gender enrichment failed")
		}
	}()

	go func() {
		defer wg.Done()
		if nationality, err := s.client.GetNationality(ctx, person.Name); err == nil {
			person.Nationality = nationality
		} else {
			s.log.WithError(err).Warn("Nationality enrichment failed")
		}
	}()

	wg.Wait()

	id, err := s.repo.Create(ctx, person)
	if err != nil {
		s.log.Errorf("Failed to create person with name %s: %v", person.Name, err)
		return 0, fmt.Errorf("failed to create person: %w", err)
	}
	return id, nil
}

func (s *PersonService) GetById(ctx context.Context, id int64) (*domain.Person, error) {
	s.log.Debugf("Getting person by ID: %d", id)
	person, err := s.repo.GetById(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.Warnf("Person with ID %d not found", id)
			return nil, err
		}
		s.log.Errorf("Failed to get person by ID %d: %v", id, err)
		return nil, fmt.Errorf("failed to get person: %w", err)
	}
	return person, nil
}

func (s *PersonService) GetAll(ctx context.Context, filters map[string]interface{}, page, pageSize int) ([]*domain.Person, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize < 1 {
		pageSize = 10
	}
	limit := pageSize
	offset := (page - 1) * pageSize

	s.log.WithFields(logrus.Fields{
		"filters":   filters,
		"page":      page,
		"page_size": pageSize,
	}).Debug("Getting people")
	people, total, err := s.repo.GetAll(ctx, filters, limit, offset)
	if err != nil {
		s.log.WithError(err).Error("Failed to get people")
		return nil, 0, fmt.Errorf("failed to get people: %w", err)
	}
	s.log.WithFields(logrus.Fields{
		"count": len(people),
		"total": total,
	}).Debug("Retrieved people")
	return people, total, nil
}

func (s *PersonService) Update(ctx context.Context, id int64, person *domain.Person) error {
	s.log.Debugf("Updating person ID: %d", id)

	err := s.repo.Update(ctx, id, person)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.Warnf("Person with ID %d not found", id)
			return err
		}
		s.log.Errorf("Failed to update person ID %d: %v", id, err)
		return fmt.Errorf("failed to update person: %w", err)
	}
	return nil
}

func (s *PersonService) Delete(ctx context.Context, id int64) error {
	s.log.Debugf("Deleting person ID %d", id)
	err := s.repo.Delete(ctx, id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			s.log.Warnf("Person with ID %d not found", id)
			return err
		}
		s.log.Errorf("Failed to delete person ID %d: %v", id, err)
		return fmt.Errorf("failed to delete person: %w", err)
	}
	return nil
}
