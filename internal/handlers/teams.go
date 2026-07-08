package handlers

import (
	"net/http"

	"github.com/go-chi/chi/v5"

	"github.com/mondial7/smart-360/internal/models"
)

// TeamDirectory lists every member grouped for a company-wide view.
func (h *Handlers) TeamDirectory(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	users, err := h.Repos.Users.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	teams, err := h.Repos.Teams.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	teamName := map[string]string{}
	for _, t := range teams {
		teamName[t.ID] = t.Name
	}
	type row struct {
		User models.User
		Team string
	}
	rows := make([]row, 0, len(users))
	for _, usr := range users {
		name := "—"
		if usr.TeamID != nil {
			if n, ok := teamName[*usr.TeamID]; ok {
				name = n
			}
		}
		rows = append(rows, row{User: usr, Team: name})
	}
	h.View.Page(w, http.StatusOK, h.page(r, "Team", "team", "team_directory_content", map[string]any{"Rows": rows}))
}

// TeamsList shows all teams (admin).
func (h *Handlers) TeamsList(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	teams, err := h.Repos.Teams.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	users, err := h.allUsersIndex(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	type card struct {
		Team      models.Team
		AdminName string
		Members   int
	}
	cards := make([]card, 0, len(teams))
	for _, t := range teams {
		cards = append(cards, card{Team: t, AdminName: users[t.TeamAdminID].Name, Members: len(t.MemberIDs)})
	}
	h.View.Page(w, http.StatusOK, h.page(r, "Teams", "teams", "teams_content", map[string]any{"Teams": cards}))
}

// NewTeamForm renders the team creation form.
func (h *Handlers) NewTeamForm(w http.ResponseWriter, r *http.Request) {
	users, err := h.Repos.Users.FindAll(r.Context())
	if err != nil {
		serverError(w, err)
		return
	}
	h.View.Page(w, http.StatusOK, h.page(r, "New team", "teams", "team_new_content", map[string]any{"Users": users}))
}

// CreateTeam creates a team with the chosen admin and members.
func (h *Handlers) CreateTeam(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	if err := r.ParseForm(); err != nil {
		http.Error(w, "bad form", http.StatusBadRequest)
		return
	}
	name := r.FormValue("name")
	adminID := r.FormValue("team_admin_id")
	if name == "" || adminID == "" {
		http.Error(w, "Name and a team admin are required", http.StatusBadRequest)
		return
	}
	team := &models.Team{Name: name, TeamAdminID: adminID}
	if err := h.Repos.Teams.Create(ctx, team); err != nil {
		serverError(w, err)
		return
	}
	// The team admin is implicitly a member.
	_ = h.Repos.Teams.AddMember(ctx, team.ID, adminID)
	if u.Role == models.RoleAdmin {
		_ = h.Repos.Users.UpdateRole(ctx, adminID, models.RoleTeamAdmin)
	}
	for _, memberID := range r.Form["member_ids"] {
		if memberID != "" && memberID != adminID {
			_ = h.Repos.Teams.AddMember(ctx, team.ID, memberID)
		}
	}
	h.audit(ctx, auditParams{Action: models.AuditTeamCreated, Actor: u, TeamID: team.ID, TeamName: name,
		Description: "Created team"})
	redirect(w, r, "/teams/"+team.ID)
}

// TeamDetails shows a team with its members and an add-member form.
func (h *Handlers) TeamDetails(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	id := chi.URLParam(r, "id")
	team, err := h.Repos.Teams.FindByID(ctx, id)
	if err != nil {
		notFound(w)
		return
	}
	users, err := h.Repos.Users.FindAll(ctx)
	if err != nil {
		serverError(w, err)
		return
	}
	index := userMap(users)
	members := make([]models.User, 0, len(team.MemberIDs))
	memberSet := map[string]bool{}
	for _, mid := range team.MemberIDs {
		memberSet[mid] = true
		members = append(members, index[mid])
	}
	// Candidates for adding: users not already members.
	var candidates []models.User
	for _, usr := range users {
		if !memberSet[usr.ID] {
			candidates = append(candidates, usr)
		}
	}
	data := map[string]any{
		"Team":       team,
		"AdminName":  index[team.TeamAdminID].Name,
		"Members":    members,
		"Candidates": candidates,
	}
	h.View.Page(w, http.StatusOK, h.page(r, team.Name, "teams", "team_details_content", data))
}

// AddTeamMember adds a member to a team.
func (h *Handlers) AddTeamMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	id := chi.URLParam(r, "id")
	_ = r.ParseForm()
	userID := r.FormValue("user_id")
	if userID != "" {
		_ = h.Repos.Teams.AddMember(ctx, id, userID)
		team, _ := h.Repos.Teams.FindByID(ctx, id)
		h.audit(ctx, auditParams{Action: models.AuditTeamMemberAdded, Actor: u, TeamID: id,
			TeamName: teamNameOf(team), Description: "Added team member"})
	}
	redirect(w, r, "/teams/"+id)
}

// RemoveTeamMember removes a member from a team.
func (h *Handlers) RemoveTeamMember(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	u := h.user(r)
	id := chi.URLParam(r, "id")
	userID := chi.URLParam(r, "userId")
	_ = h.Repos.Teams.RemoveMember(ctx, id, userID)
	team, _ := h.Repos.Teams.FindByID(ctx, id)
	h.audit(ctx, auditParams{Action: models.AuditTeamMemberRemoved, Actor: u, TeamID: id,
		TeamName: teamNameOf(team), Description: "Removed team member"})
	redirect(w, r, "/teams/"+id)
}

func teamNameOf(t *models.Team) string {
	if t == nil {
		return ""
	}
	return t.Name
}
