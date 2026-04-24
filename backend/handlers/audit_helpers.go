package handlers

import (
	"context"
	"fmt"
	"smart360/database"
	"smart360/models"
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

// AuditLogParams contains the parameters for creating an audit log entry
type AuditLogParams struct {
	Action       models.AuditAction
	ActorID      primitive.ObjectID
	ActorName    string
	ActorEmail   string
	RoundID      primitive.ObjectID
	RoundSubject string
	TeamID       primitive.ObjectID
	TeamName     string
	Description  string
	OldValue     string
	NewValue     string
	Metadata     string
}

// createAuditLog inserts an audit log entry into the database
// Errors are logged but do not fail the main operation
func createAuditLog(ctx context.Context, params AuditLogParams) {
	db := database.GetDB()

	auditLog := models.AuditLog{
		ID:           primitive.NewObjectID(),
		Action:       params.Action,
		ActorID:      params.ActorID,
		ActorName:    params.ActorName,
		ActorEmail:   params.ActorEmail,
		RoundID:      params.RoundID,
		RoundSubject: params.RoundSubject,
		TeamID:       params.TeamID,
		TeamName:     params.TeamName,
		Description:  params.Description,
		OldValue:     params.OldValue,
		NewValue:     params.NewValue,
		Metadata:     params.Metadata,
		CreatedAt:    time.Now(),
	}

	_, err := db.Collection("audit_logs").InsertOne(ctx, auditLog)
	if err != nil {
		// Log error but don't fail the main operation
		fmt.Printf("Failed to create audit log: %v\n", err)
	}
}
