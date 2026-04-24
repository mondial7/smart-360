package handlers

import (
	"context"
	"fmt"
	"smart360/models"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

// PopulatedTeam includes user details for team members and admin
type PopulatedTeam struct {
	ID          primitive.ObjectID   `json:"id"`
	Name        string               `json:"name"`
	TeamAdminID primitive.ObjectID   `json:"teamAdminId"`
	TeamAdmin   *models.User         `json:"teamAdmin,omitempty"`
	MemberIDs   []primitive.ObjectID `json:"memberIds"`
	Members     []models.User        `json:"members"`
	CreatedAt   string               `json:"createdAt"`
	UpdatedAt   string               `json:"updatedAt"`
}

// getPopulatedTeam fetches a team and populates user details
func getPopulatedTeam(ctx context.Context, db *mongo.Database, teamID primitive.ObjectID) (*PopulatedTeam, error) {
	var team models.Team
	err := db.Collection("teams").FindOne(ctx, bson.M{"_id": teamID}).Decode(&team)
	if err != nil {
		return nil, err
	}

	populated := &PopulatedTeam{
		ID:          team.ID,
		Name:        team.Name,
		TeamAdminID: team.TeamAdminID,
		MemberIDs:   team.MemberIDs,
		Members:     []models.User{},
		CreatedAt:   team.CreatedAt.Format("2006-01-02T15:04:05Z"),
		UpdatedAt:   team.UpdatedAt.Format("2006-01-02T15:04:05Z"),
	}

	// Fetch team admin
	var admin models.User
	err = db.Collection("users").FindOne(ctx, bson.M{"_id": team.TeamAdminID}).Decode(&admin)
	if err == nil {
		populated.TeamAdmin = &admin
	}

	// Fetch all team members
	cursor, err := db.Collection("users").Find(ctx, bson.M{"_id": bson.M{"$in": team.MemberIDs}})
	if err == nil {
		cursor.All(ctx, &populated.Members)
	}

	return populated, nil
}

// validateUserNotInTeam checks if a user is not already assigned to a team
func validateUserNotInTeam(ctx context.Context, db *mongo.Database, userID primitive.ObjectID) error {
	var user models.User
	err := db.Collection("users").FindOne(ctx, bson.M{"_id": userID}).Decode(&user)
	if err != nil {
		return fmt.Errorf("user not found")
	}

	if user.TeamID != nil {
		return fmt.Errorf("user already assigned to a team")
	}

	return nil
}

// updateUserTeamAssignments sets the team_id for multiple users
func updateUserTeamAssignments(ctx context.Context, db *mongo.Database, userIDs []primitive.ObjectID, teamID primitive.ObjectID) error {
	for _, userID := range userIDs {
		_, err := db.Collection("users").UpdateOne(
			ctx,
			bson.M{"_id": userID},
			bson.M{"$set": bson.M{"team_id": teamID}},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// clearUserTeamAssignments removes the team_id for multiple users
func clearUserTeamAssignments(ctx context.Context, db *mongo.Database, userIDs []primitive.ObjectID) error {
	for _, userID := range userIDs {
		_, err := db.Collection("users").UpdateOne(
			ctx,
			bson.M{"_id": userID},
			bson.M{"$unset": bson.M{"team_id": ""}},
		)
		if err != nil {
			return err
		}
	}
	return nil
}

// canManageTeam checks if a user can manage the given team
func canManageTeam(user models.User, teamID primitive.ObjectID) bool {
	// Global admin can manage all teams
	if user.Role == models.RoleAdmin {
		return true
	}

	// Team admin can manage their own team
	if user.Role == models.RoleTeamAdmin && user.TeamID != nil && *user.TeamID == teamID {
		return true
	}

	return false
}

// contains checks if a slice contains a specific ObjectID
func contains(slice []primitive.ObjectID, item primitive.ObjectID) bool {
	for _, s := range slice {
		if s == item {
			return true
		}
	}
	return false
}
