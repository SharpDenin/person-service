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
