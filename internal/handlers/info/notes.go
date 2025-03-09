// SPDX-License-Identifier: BSD-2-Clause
//
// Copyright (c) 2025 The FreeBSD Foundation.
//
// This software was developed by Hayzam Sherif <hayzam@alchemilla.io>
// of Alchemilla Ventures Pvt. Ltd. <hello@alchemilla.io>,
// under sponsorship from the FreeBSD Foundation.

package infoHandlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"sylve/internal/services/info"

	"github.com/gin-gonic/gin"
)

type Response struct {
	Status string      `json:"status"`
	Data   interface{} `json:"data,omitempty"`
	Error  string      `json:"error,omitempty"`
}

type NoteRequest struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

func validateNoteRequest(note NoteRequest) error {
	return Validate.StructPartial(note, "Title", "Content")
}

func getNoteID(c *gin.Context) (int, error) {
	idStr := c.Param("id")
	id, err := strconv.Atoi(idStr)
	if err != nil {
		return 0, err
	}
	return id, nil
}

func NotesHandler(infoService *info.Service) gin.HandlerFunc {
	return func(c *gin.Context) {
		switch c.Request.Method {
		case http.MethodGet:
			handleGetNotes(c, infoService)
		case http.MethodPost:
			handlePostNotes(c, infoService)
		case http.MethodDelete:
			handleDeleteNoteByID(c, infoService)
		case http.MethodPut:
			handleUpdateNoteByID(c, infoService)
		default:
			c.JSON(http.StatusMethodNotAllowed, Response{
				Status: "error",
				Error:  "Method not allowed",
			})
		}
	}
}

func handleGetNotes(c *gin.Context, infoService *info.Service) {
	notes, err := infoService.GetNotes()
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Status: "success",
		Data:   notes,
	})
}

func handlePostNotes(c *gin.Context, infoService *info.Service) {
	decoder := json.NewDecoder(c.Request.Body)
	decoder.DisallowUnknownFields()

	var newNote NoteRequest
	if err := decoder.Decode(&newNote); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  "Invalid request payload, unknown fields are not allowed",
		})
		return
	}

	if err := validateNoteRequest(newNote); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	err := infoService.AddNote(newNote.Title, newNote.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusCreated, Response{
		Status: "success",
	})
}

func handleDeleteNoteByID(c *gin.Context, infoService *info.Service) {
	id, err := getNoteID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  "Invalid note ID",
		})
		return
	}

	err = infoService.DeleteNoteByID(id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}
	c.JSON(http.StatusOK, Response{
		Status: "success",
	})
}

func handleUpdateNoteByID(c *gin.Context, infoService *info.Service) {
	id, err := getNoteID(c)
	if err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  "Invalid note ID",
		})
		return
	}

	var updateData NoteRequest
	if err := c.ShouldBindJSON(&updateData); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  "Invalid request payload",
		})
		return
	}

	if err := validateNoteRequest(updateData); err != nil {
		c.JSON(http.StatusBadRequest, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	err = infoService.UpdateNoteByID(id, updateData.Title, updateData.Content)
	if err != nil {
		c.JSON(http.StatusInternalServerError, Response{
			Status: "error",
			Error:  err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, Response{
		Status: "success",
	})
}
