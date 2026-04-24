package handlers

import (
	"context"
	"fmt"
	"net/http"
	"smart360/database"
	"smart360/models"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// GetTeams returns all teams with populated user details (admin only)
func GetTeams(c *gin.Context) {
	db := database.GetDB()
	ctx := context.Background()

	cursor, err := db.Collection("teams").Find(ctx, bson.M{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to fetch teams"})
		return
	}

	var teams []models.Team
	if err := cursor.All(ctx, &teams); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to decode teams"})
		return
	}

	// Populate teams with user details
	populatedTeams := make([]PopulatedTeam, 0, len(teams))
	for _, team := range teams {
		populated, err := getPopulatedTeam(ctx, db, team.ID)
		if err != nil {
			continue
		}
		populatedTeams = append(populatedTeams, *populated)
	}

	c.JSON(http.StatusOK, populatedTeams)
}

// GetTeam returns a single team with populated user details
func GetTeam(c *gin.Context) {
	teamID := c.Param("id")
	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	populated, err := getPopulatedTeam(ctx, db, teamObjID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	c.JSON(http.StatusOK, populated)
}

// GetMyTeam returns the current user's team
func GetMyTeam(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	if currentUser.TeamID == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "You are not assigned to a team"})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	populated, err := getPopulatedTeam(ctx, db, *currentUser.TeamID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	c.JSON(http.StatusOK, populated)
}

// CreateTeam creates a new team (admin only)
func CreateTeam(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	var req struct {
		Name        string               `json:"name" binding:"required"`
		TeamAdminID primitive.ObjectID   `json:"teamAdminId" binding:"required"`
		MemberIDs   []primitive.ObjectID `json:"memberIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	// Validate team admin exists
	var admin models.User
	err := db.Collection("users").FindOne(ctx, bson.M{"_id": req.TeamAdminID}).Decode(&admin)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team admin user not found"})
		return
	}

	// Validate team admin is in member list
	if !contains(req.MemberIDs, req.TeamAdminID) {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Team admin must be included in team members"})
		return
	}

	// Validate all members exist and are not in another team
	for _, memberID := range req.MemberIDs {
		if err := validateUserNotInTeam(ctx, db, memberID); err != nil {
			var member models.User
			db.Collection("users").FindOne(ctx, bson.M{"_id": memberID}).Decode(&member)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("User %s is already assigned to a team", member.Name)})
			return
		}
	}

	// Create the team
	now := time.Now()
	team := models.Team{
		Name:        req.Name,
		TeamAdminID: req.TeamAdminID,
		MemberIDs:   req.MemberIDs,
		CreatedAt:   now,
		UpdatedAt:   now,
	}

	result, err := db.Collection("teams").InsertOne(ctx, team)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create team"})
		return
	}

	teamID := result.InsertedID.(primitive.ObjectID)
	team.ID = teamID

	// Update all member TeamID references
	if err := updateUserTeamAssignments(ctx, db, req.MemberIDs, teamID); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to assign members to team"})
		return
	}

	// Promote team admin role
	_, err = db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": req.TeamAdminID},
		bson.M{"$set": bson.M{"role": models.RoleTeamAdmin}},
	)
	if err != nil {
		fmt.Printf("Failed to promote team admin: %v\n", err)
	}

	// Create audit log
	createAuditLog(ctx, AuditLogParams{
		Action:      models.AuditTeamCreated,
		ActorID:     currentUser.ID,
		ActorName:   currentUser.Name,
		ActorEmail:  currentUser.Email,
		TeamID:      teamID,
		TeamName:    req.Name,
		Description: fmt.Sprintf("Created team '%s' with %d members", req.Name, len(req.MemberIDs)),
		Metadata:    fmt.Sprintf(`{"team_admin": "%s", "member_count": %d}`, admin.Name, len(req.MemberIDs)),
	})

	// Return populated team
	populated, _ := getPopulatedTeam(ctx, db, teamID)
	c.JSON(http.StatusCreated, populated)
}

// UpdateTeam updates team details
func UpdateTeam(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	teamID := c.Param("id")
	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req struct {
		Name        *string               `json:"name"`
		TeamAdminID *primitive.ObjectID   `json:"teamAdminId"`
		MemberIDs   *[]primitive.ObjectID `json:"memberIds"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	// Get existing team
	var team models.Team
	err = db.Collection("teams").FindOne(ctx, bson.M{"_id": teamObjID}).Decode(&team)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Check authorization
	if !canManageTeam(currentUser, teamObjID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to update this team"})
		return
	}

	updates := bson.M{"updated_at": time.Now()}

	// Update team name
	if req.Name != nil && *req.Name != "" {
		updates["name"] = *req.Name
	}

	// Handle team admin change
	if req.TeamAdminID != nil && *req.TeamAdminID != team.TeamAdminID {
		oldAdminID := team.TeamAdminID
		newAdminID := *req.TeamAdminID

		// Validate new admin is a team member
		if req.MemberIDs != nil {
			if !contains(*req.MemberIDs, newAdminID) {
				c.JSON(http.StatusBadRequest, gin.H{"error": "New team admin must be a team member"})
				return
			}
		} else if !contains(team.MemberIDs, newAdminID) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "New team admin must be a team member"})
			return
		}

		// Demote old admin
		db.Collection("users").UpdateOne(
			ctx,
			bson.M{"_id": oldAdminID},
			bson.M{"$set": bson.M{"role": models.RoleMember}},
		)

		// Promote new admin
		db.Collection("users").UpdateOne(
			ctx,
			bson.M{"_id": newAdminID},
			bson.M{"$set": bson.M{"role": models.RoleTeamAdmin}},
		)

		updates["team_admin_id"] = newAdminID

		// Audit log for admin change
		var oldAdmin, newAdmin models.User
		db.Collection("users").FindOne(ctx, bson.M{"_id": oldAdminID}).Decode(&oldAdmin)
		db.Collection("users").FindOne(ctx, bson.M{"_id": newAdminID}).Decode(&newAdmin)

		createAuditLog(ctx, AuditLogParams{
			Action:      models.AuditTeamAdminChanged,
			ActorID:     currentUser.ID,
			ActorName:   currentUser.Name,
			ActorEmail:  currentUser.Email,
			TeamID:      teamObjID,
			TeamName:    team.Name,
			Description: fmt.Sprintf("Changed team admin from %s to %s", oldAdmin.Name, newAdmin.Name),
			OldValue:    oldAdmin.Name,
			NewValue:    newAdmin.Name,
		})
	}

	// Handle member changes
	if req.MemberIDs != nil {
		// Calculate additions and removals
		oldMembers := team.MemberIDs
		newMembers := *req.MemberIDs

		// Find removed members
		for _, oldMember := range oldMembers {
			if !contains(newMembers, oldMember) {
				// Check if it's the team admin
				if oldMember == team.TeamAdminID {
					c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove team admin from members"})
					return
				}

				// Clear team assignment
				clearUserTeamAssignments(ctx, db, []primitive.ObjectID{oldMember})

				// Audit log
				var user models.User
				db.Collection("users").FindOne(ctx, bson.M{"_id": oldMember}).Decode(&user)
				createAuditLog(ctx, AuditLogParams{
					Action:      models.AuditTeamMemberRemoved,
					ActorID:     currentUser.ID,
					ActorName:   currentUser.Name,
					ActorEmail:  currentUser.Email,
					TeamID:      teamObjID,
					TeamName:    team.Name,
					Description: fmt.Sprintf("Removed %s from team", user.Name),
					OldValue:    user.Name,
				})
			}
		}

		// Find added members
		for _, newMember := range newMembers {
			if !contains(oldMembers, newMember) {
				// Validate not in another team
				if err := validateUserNotInTeam(ctx, db, newMember); err != nil {
					var user models.User
					db.Collection("users").FindOne(ctx, bson.M{"_id": newMember}).Decode(&user)
					c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("User %s is already in another team", user.Name)})
					return
				}

				// Assign to team
				updateUserTeamAssignments(ctx, db, []primitive.ObjectID{newMember}, teamObjID)

				// Audit log
				var user models.User
				db.Collection("users").FindOne(ctx, bson.M{"_id": newMember}).Decode(&user)
				createAuditLog(ctx, AuditLogParams{
					Action:      models.AuditTeamMemberAdded,
					ActorID:     currentUser.ID,
					ActorName:   currentUser.Name,
					ActorEmail:  currentUser.Email,
					TeamID:      teamObjID,
					TeamName:    team.Name,
					Description: fmt.Sprintf("Added %s to team", user.Name),
					NewValue:    user.Name,
				})
			}
		}

		updates["member_ids"] = newMembers
	}

	// Apply updates
	_, err = db.Collection("teams").UpdateOne(
		ctx,
		bson.M{"_id": teamObjID},
		bson.M{"$set": updates},
	)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update team"})
		return
	}

	// Return updated team
	populated, _ := getPopulatedTeam(ctx, db, teamObjID)
	c.JSON(http.StatusOK, populated)
}

// DeleteTeam deletes a team (admin only)
func DeleteTeam(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	teamID := c.Param("id")
	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	// Get team
	var team models.Team
	err = db.Collection("teams").FindOne(ctx, bson.M{"_id": teamObjID}).Decode(&team)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Check for active/draft rounds
	roundCount, err := db.Collection("feedback_rounds").CountDocuments(ctx, bson.M{
		"subject_id": bson.M{"$in": team.MemberIDs},
		"status":     bson.M{"$in": []string{"draft", "active"}},
	})
	if err == nil && roundCount > 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot delete team with active or draft feedback rounds"})
		return
	}

	// Clear all member team assignments
	clearUserTeamAssignments(ctx, db, team.MemberIDs)

	// Demote team admin
	db.Collection("users").UpdateOne(
		ctx,
		bson.M{"_id": team.TeamAdminID},
		bson.M{"$set": bson.M{"role": models.RoleMember}},
	)

	// Delete team
	_, err = db.Collection("teams").DeleteOne(ctx, bson.M{"_id": teamObjID})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to delete team"})
		return
	}

	// Audit log
	createAuditLog(ctx, AuditLogParams{
		Action:      models.AuditTeamDeleted,
		ActorID:     currentUser.ID,
		ActorName:   currentUser.Name,
		ActorEmail:  currentUser.Email,
		TeamID:      teamObjID,
		TeamName:    team.Name,
		Description: fmt.Sprintf("Deleted team '%s'", team.Name),
	})

	c.JSON(http.StatusOK, gin.H{"message": "Team deleted successfully"})
}

// AddTeamMembers adds members to a team
func AddTeamMembers(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	teamID := c.Param("id")
	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	var req struct {
		MemberIDs []primitive.ObjectID `json:"memberIds" binding:"required"`
	}

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	// Get team
	var team models.Team
	err = db.Collection("teams").FindOne(ctx, bson.M{"_id": teamObjID}).Decode(&team)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Check authorization
	if !canManageTeam(currentUser, teamObjID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to manage this team"})
		return
	}

	// Validate and add members
	for _, memberID := range req.MemberIDs {
		// Skip if already in team
		if contains(team.MemberIDs, memberID) {
			continue
		}

		// Validate not in another team
		if err := validateUserNotInTeam(ctx, db, memberID); err != nil {
			var user models.User
			db.Collection("users").FindOne(ctx, bson.M{"_id": memberID}).Decode(&user)
			c.JSON(http.StatusBadRequest, gin.H{"error": fmt.Sprintf("User %s is already in another team", user.Name)})
			return
		}

		// Add to team
		db.Collection("teams").UpdateOne(
			ctx,
			bson.M{"_id": teamObjID},
			bson.M{"$push": bson.M{"member_ids": memberID}},
		)

		// Update user
		updateUserTeamAssignments(ctx, db, []primitive.ObjectID{memberID}, teamObjID)

		// Audit log
		var user models.User
		db.Collection("users").FindOne(ctx, bson.M{"_id": memberID}).Decode(&user)
		createAuditLog(ctx, AuditLogParams{
			Action:      models.AuditTeamMemberAdded,
			ActorID:     currentUser.ID,
			ActorName:   currentUser.Name,
			ActorEmail:  currentUser.Email,
			TeamID:      teamObjID,
			TeamName:    team.Name,
			Description: fmt.Sprintf("Added %s to team", user.Name),
			NewValue:    user.Name,
		})
	}

	// Return updated team
	populated, _ := getPopulatedTeam(ctx, db, teamObjID)
	c.JSON(http.StatusOK, populated)
}

// RemoveTeamMember removes a member from a team
func RemoveTeamMember(c *gin.Context) {
	user, _ := c.Get("user")
	currentUser := user.(models.User)

	teamID := c.Param("id")
	memberID := c.Param("userId")

	teamObjID, err := primitive.ObjectIDFromHex(teamID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid team ID"})
		return
	}

	memberObjID, err := primitive.ObjectIDFromHex(memberID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID"})
		return
	}

	db := database.GetDB()
	ctx := context.Background()

	// Get team
	var team models.Team
	err = db.Collection("teams").FindOne(ctx, bson.M{"_id": teamObjID}).Decode(&team)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Team not found"})
		return
	}

	// Check authorization
	if !canManageTeam(currentUser, teamObjID) {
		c.JSON(http.StatusForbidden, gin.H{"error": "Not authorized to manage this team"})
		return
	}

	// Cannot remove team admin
	if memberObjID == team.TeamAdminID {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Cannot remove team admin. Change team admin first."})
		return
	}

	// Remove from team
	db.Collection("teams").UpdateOne(
		ctx,
		bson.M{"_id": teamObjID},
		bson.M{"$pull": bson.M{"member_ids": memberObjID}},
	)

	// Clear user team assignment
	clearUserTeamAssignments(ctx, db, []primitive.ObjectID{memberObjID})

	// Audit log
	var removedUser models.User
	db.Collection("users").FindOne(ctx, bson.M{"_id": memberObjID}).Decode(&removedUser)
	createAuditLog(ctx, AuditLogParams{
		Action:      models.AuditTeamMemberRemoved,
		ActorID:     currentUser.ID,
		ActorName:   currentUser.Name,
		ActorEmail:  currentUser.Email,
		TeamID:      teamObjID,
		TeamName:    team.Name,
		Description: fmt.Sprintf("Removed %s from team", removedUser.Name),
		OldValue:    removedUser.Name,
	})

	// Return updated team
	populated, _ := getPopulatedTeam(ctx, db, teamObjID)
	c.JSON(http.StatusOK, populated)
}
