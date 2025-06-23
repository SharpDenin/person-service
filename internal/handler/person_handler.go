package handler

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"net/http"
	"os"
	"person-service/internal/domain"
	"person-service/internal/repository"
	"person-service/internal/service"
	"strconv"
	"time"
)

const (
	paginationMaxPageSize     = 100
	paginationDefaultPageSize = 10
	paginationDefaultPage     = 1
)

type PersonHandler struct {
	service service.PersonServiceInterface
	log     *logrus.Logger
}

func NewPersonHandler(service service.PersonServiceInterface, log *logrus.Logger) *PersonHandler {
	if log == nil {
		log = logrus.New()
		log.SetFormatter(&logrus.JSONFormatter{})
		log.SetOutput(os.Stdout)
		log.SetLevel(logrus.DebugLevel)
	}
	return &PersonHandler{
		service: service,
		log:     log,
	}
}

// CreatePerson creates a new person
// @Summary Create a new person
// @Description Creates a person with enriched age, gender, and nationality from external APIs
// @Tags persons
// @Accept json
// @Produce json
// @Param person body domain.CreatePersonRequest true "Person data"
// @Success 201 {object} domain.Person
// @Failure 400 {object} domain.ErrorResponse "Invalid request body"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /person [post]
func (h *PersonHandler) CreatePerson(c *gin.Context) {
	var request domain.CreatePersonRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.WithError(err).Debug("Failed to bind request")
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid request body"})
		return
	}

	var patronymicPtr *string
	if request.Patronymic != "" {
		patronymicPtr = &request.Patronymic
	}

	person := &domain.Person{
		Name:       request.Name,
		Surname:    request.Surname,
		Patronymic: patronymicPtr,
		CreatedAt:  time.Now().UTC(),
	}

	id, err := h.service.Create(c.Request.Context(), person)
	if err != nil {
		h.log.WithFields(logrus.Fields{
			"error": err,
			"name":  request.Name,
		}).Error("Failed to create person")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to create person"})
		return
	}

	person.ID = id
	h.log.WithField("id", id).Info("Person created successfully")
	c.JSON(http.StatusCreated, person)
}

// GetPerson retrieves a person by ID
// @Summary Get a person by ID
// @Description Retrieves a person by their unique ID
// @Tags persons
// @Produce json
// @Param id path int true "Person ID"
// @Success 200 {object} domain.Person
// @Failure 400 {object} domain.ErrorResponse "Invalid ID format"
// @Failure 404 {object} domain.ErrorResponse "Person not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /person/{id} [get]
func (h *PersonHandler) GetPerson(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.WithError(err).Debug("Invalid ID parameter")
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid ID format"})
		return
	}

	person, err := h.service.GetById(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.log.WithField("id", id).Debug("Person not found")
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: "Person not found"})
			return
		}
		h.log.WithError(err).Error("Failed to get person")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to get person"})
		return
	}

	h.log.WithField("id", id).Info("Person retrieved successfully")
	c.JSON(http.StatusOK, person)
}

// GetAll retrieves all persons with pagination and filtering
// @Summary Get all persons
// @Description Retrieves a list of persons with optional filters and pagination
// @Tags persons
// @Produce json
// @Param page query int false "Page number" default(1)
// @Param page_size query int false "Page size" default(10)
// @Param name query string false "Filter by name"
// @Param surname query string false "Filter by surname"
// @Param age query int false "Filter by age"
// @Param gender query string false "Filter by gender"
// @Param nationality query string false "Filter by nationality"
// @Success 200 {object} domain.PersonListResponse
// @Failure 400 {object} domain.ErrorResponse "Invalid query parameters"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /people [get]
func (h *PersonHandler) GetAll(c *gin.Context) {
	page, err := strconv.Atoi(c.DefaultQuery("page", strconv.Itoa(paginationDefaultPage)))
	if err != nil || page < 1 {
		page = paginationDefaultPage
	}

	pageSize, err := strconv.Atoi(c.DefaultQuery("page_size", strconv.Itoa(paginationDefaultPageSize)))
	if err != nil || pageSize < 1 {
		pageSize = paginationDefaultPageSize
	}
	if pageSize > paginationMaxPageSize {
		pageSize = paginationMaxPageSize
	}

	filters := make(map[string]interface{})
	if name := c.Query("name"); name != "" {
		filters["name"] = name
	}
	if surname := c.Query("surname"); surname != "" {
		filters["surname"] = surname
	}
	if ageStr := c.Query("age"); ageStr != "" {
		if age, err := strconv.Atoi(ageStr); err == nil {
			filters["age"] = age
		} else {
			h.log.WithField("age", ageStr).Debug("Invalid age parameter")
			c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid age format"})
			return
		}
	}
	if gender := c.Query("gender"); gender != "" {
		filters["gender"] = gender
	}
	if nationality := c.Query("nationality"); nationality != "" {
		filters["nationality"] = nationality
	}

	persons, total, err := h.service.GetAll(c.Request.Context(), filters, page, pageSize)
	if err != nil {
		h.log.WithError(err).Error("Failed to get persons")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to get persons"})
		return
	}

	h.log.WithFields(logrus.Fields{
		"count":     len(persons),
		"total":     total,
		"page":      page,
		"page_size": pageSize,
	}).Info("Persons retrieved successfully")

	response := domain.PersonListResponse{
		Data: persons,
		Meta: domain.PaginationMeta{
			Page:       page,
			PageSize:   pageSize,
			TotalItems: total,
		},
	}

	c.JSON(http.StatusOK, response)
}

// Update updates a person by ID
// @Summary Update a person
// @Description Updates a person's details by their ID
// @Tags persons
// @Accept json
// @Produce json
// @Param id path int true "Person ID"
// @Param person body domain.UpdatePersonRequest true "Person data"
// @Success 200 {object} domain.Person
// @Failure 400 {object} domain.ErrorResponse "Invalid request body or ID"
// @Failure 404 {object} domain.ErrorResponse "Person not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /person/{id} [put]
func (h *PersonHandler) Update(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.WithError(err).Debug("Invalid ID parameter")
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid ID format"})
		return
	}

	var request domain.UpdatePersonRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		h.log.WithError(err).Debug("Failed to bind request")
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid request body"})
		return
	}

	var patronymicPtr *string
	if request.Patronymic != "" {
		patronymicPtr = &request.Patronymic
	}

	person := &domain.Person{
		Name:        request.Name,
		Surname:     request.Surname,
		Patronymic:  patronymicPtr,
		Age:         request.Age,
		Gender:      request.Gender,
		Nationality: request.Nationality,
	}

	if err := h.service.Update(c.Request.Context(), id, person); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.log.WithField("id", id).Debug("Person not found for update")
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: "Person not found"})
			return
		}
		h.log.WithError(err).Error("Failed to update person")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to update person"})
		return
	}

	h.log.WithField("id", id).Info("Person updated successfully")
	c.JSON(http.StatusOK, person)
}

// Delete deletes a person by ID
// @Summary Delete a person
// @Description Deletes a person by their ID
// @Tags persons
// @Produce json
// @Param id path int true "Person ID"
// @Success 204
// @Failure 400 {object} domain.ErrorResponse "Invalid ID format"
// @Failure 404 {object} domain.ErrorResponse "Person not found"
// @Failure 500 {object} domain.ErrorResponse "Internal server error"
// @Router /person/{id} [delete]
func (h *PersonHandler) Delete(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		h.log.WithError(err).Debug("Invalid ID parameter")
		c.JSON(http.StatusBadRequest, domain.ErrorResponse{Error: "Invalid ID format"})
		return
	}

	if err := h.service.Delete(c.Request.Context(), id); err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			h.log.WithField("id", id).Debug("Person not found for deletion")
			c.JSON(http.StatusNotFound, domain.ErrorResponse{Error: "Person not found"})
			return
		}
		h.log.WithError(err).Error("Failed to delete person")
		c.JSON(http.StatusInternalServerError, domain.ErrorResponse{Error: "Failed to delete person"})
		return
	}

	h.log.WithField("id", id).Info("Person deleted successfully")
	c.Status(http.StatusNoContent)
}
