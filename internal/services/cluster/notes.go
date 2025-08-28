package cluster

import (
	"encoding/json"
	"fmt"
	"time"

	clusterModels "github.com/alchemillahq/sylve/internal/db/models/cluster"
)

func (s *Service) ListNotes() ([]clusterModels.ClusterNote, error) {
	var notes []clusterModels.ClusterNote
	err := s.DB.Order("id ASC").Find(&notes).Error
	return notes, err
}

func (s *Service) ProposeNoteCreate(title, content string, bypassRaft bool) error {
	if bypassRaft {
		note := clusterModels.ClusterNote{
			Title:   title,
			Content: content,
		}

		return s.DB.Create(&note).Error
	}

	if s.Raft == nil {
		return fmt.Errorf("raft_not_initialized")
	}

	payloadStruct := struct {
		Title   string `json:"title"`
		Content string `json:"content"`
	}{
		Title:   title,
		Content: content,
	}

	data, err := json.Marshal(payloadStruct)
	if err != nil {
		return fmt.Errorf("failed_to_marshal_note_payload: %w", err)
	}

	cmd := clusterModels.Command{
		Type:   "note",
		Action: "create",
		Data:   data,
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed_to_marshal_command: %w", err)
	}

	applyFuture := s.Raft.Apply(payload, 5*time.Second)
	if err := applyFuture.Error(); err != nil {
		return fmt.Errorf("raft_apply_failed: %w", err)
	}

	if resp, ok := applyFuture.Response().(error); ok && resp != nil {
		return fmt.Errorf("fsm_apply_failed: %w", resp)
	}

	return nil
}

func (s *Service) ProposeNoteUpdate(id int, title, content string) error {
	if s.Raft == nil {
		return fmt.Errorf("raft_not_initialized")
	}

	payloadStruct := struct {
		ID      int    `json:"id"`
		Title   string `json:"title"`
		Content string `json:"content"`
	}{
		ID:      id,
		Title:   title,
		Content: content,
	}

	data, err := json.Marshal(payloadStruct)
	if err != nil {
		return fmt.Errorf("failed_to_marshal_note_payload: %w", err)
	}

	cmd := clusterModels.Command{
		Type:   "note",
		Action: "update",
		Data:   data,
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed_to_marshal_command: %w", err)
	}

	applyFuture := s.Raft.Apply(payload, 5*time.Second)
	if err := applyFuture.Error(); err != nil {
		return fmt.Errorf("raft_apply_failed: %w", err)
	}

	if resp, ok := applyFuture.Response().(error); ok && resp != nil {
		return fmt.Errorf("fsm_apply_failed: %w", resp)
	}

	return nil
}

func (s *Service) ProposeNoteDelete(id int) error {
	if s.Raft == nil {
		return fmt.Errorf("raft_not_initialized")
	}

	payloadStruct := struct {
		ID int `json:"id"`
	}{ID: id}

	data, err := json.Marshal(payloadStruct)
	if err != nil {
		return fmt.Errorf("failed_to_marshal_delete_payload: %w", err)
	}

	cmd := clusterModels.Command{
		Type:   "note",
		Action: "delete",
		Data:   data,
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed_to_marshal_command: %w", err)
	}

	applyFuture := s.Raft.Apply(payload, 5*time.Second)
	if err := applyFuture.Error(); err != nil {
		return fmt.Errorf("raft_apply_failed: %w", err)
	}

	if resp, ok := applyFuture.Response().(error); ok && resp != nil {
		return fmt.Errorf("fsm_apply_failed: %w", resp)
	}

	return nil
}

func (s *Service) ProposeNoteBulkDelete(ids []int) error {
	if s.Raft == nil {
		return fmt.Errorf("raft_not_initialized")
	}

	payloadStruct := struct {
		IDs []int `json:"ids"`
	}{IDs: ids}

	data, err := json.Marshal(payloadStruct)
	if err != nil {
		return fmt.Errorf("failed_to_marshal_bulk_delete_payload: %w", err)
	}

	cmd := clusterModels.Command{
		Type:   "note",
		Action: "bulk_delete",
		Data:   data,
	}

	payload, err := json.Marshal(cmd)
	if err != nil {
		return fmt.Errorf("failed_to_marshal_command: %w", err)
	}

	applyFuture := s.Raft.Apply(payload, 5*time.Second)
	if err := applyFuture.Error(); err != nil {
		return fmt.Errorf("raft_apply_failed: %w", err)
	}

	if resp, ok := applyFuture.Response().(error); ok && resp != nil {
		return fmt.Errorf("fsm_apply_failed: %w", resp)
	}

	return nil
}
