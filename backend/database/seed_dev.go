package database

import (
	"context"
	"encoding/json"
	"log"
	"smart360/models"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

// SeedDevData creates comprehensive test data for development
func SeedDevData() {
	db := GetDB()
	ctx := context.Background()

	log.Println("🌱 Seeding comprehensive development data...")

	// Clear existing data
	log.Println("🧹 Clearing existing data...")
	db.Collection("users").DeleteMany(ctx, bson.M{})
	db.Collection("feedback_rounds").DeleteMany(ctx, bson.M{})
	db.Collection("submissions").DeleteMany(ctx, bson.M{})
	db.Collection("consolidations").DeleteMany(ctx, bson.M{})

	now := time.Now()

	// Create users
	log.Println("👥 Creating users...")
	users := []models.User{
		{
			Email:     "admin@example.com",
			Name:      "Emma Admin",
			PhotoURL:  "",
			Role:      models.RoleAdmin,
			LastLogin: &now,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Email:     "alice@example.com",
			Name:      "Alice Johnson",
			PhotoURL:  "",
			Role:      models.RoleMember,
			LastLogin: &now,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Email:     "bob@example.com",
			Name:      "Bob Smith",
			PhotoURL:  "",
			Role:      models.RoleMember,
			LastLogin: &now,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Email:     "carol@example.com",
			Name:      "Carol Williams",
			PhotoURL:  "",
			Role:      models.RoleMember,
			LastLogin: &now,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Email:     "david@example.com",
			Name:      "David Brown",
			PhotoURL:  "",
			Role:      models.RoleMember,
			CreatedAt: now,
			UpdatedAt: now,
		},
		{
			Email:     "eve@example.com",
			Name:      "Eve Martinez",
			PhotoURL:  "",
			Role:      models.RoleMember,
			CreatedAt: now,
			UpdatedAt: now,
		},
	}

	userIDs := make(map[string]primitive.ObjectID)
	for _, user := range users {
		result, err := db.Collection("users").InsertOne(ctx, user)
		if err != nil {
			log.Printf("❌ Failed to seed user %s: %v", user.Email, err)
			continue
		}
		userIDs[user.Email] = result.InsertedID.(primitive.ObjectID)
		log.Printf("✅ Created user: %s", user.Name)
	}

	// Get admin and member IDs
	adminID := userIDs["admin@example.com"]
	aliceID := userIDs["alice@example.com"]
	bobID := userIDs["bob@example.com"]
	carolID := userIDs["carol@example.com"]
	davidID := userIDs["david@example.com"]
	eveID := userIDs["eve@example.com"]

	// Create feedback rounds in different statuses
	log.Println("🔄 Creating feedback rounds...")

	// 1. DRAFT ROUND - Just created, no reviewers yet
	draftRound := models.FeedbackRound{
		SubjectID:   aliceID,
		CreatedByID: adminID,
		Status:      models.RoundDraft,
		CreatedAt:   now.AddDate(0, 0, -1),
		UpdatedAt:   now.AddDate(0, 0, -1),
	}
	_, err := db.Collection("feedback_rounds").InsertOne(ctx, draftRound)
	if err != nil {
		log.Printf("❌ Failed to create draft round: %v", err)
	} else {
		log.Printf("✅ Created DRAFT round for Alice")
	}

	// 2. ACTIVE ROUND - In progress, deadline in future, partial submissions
	futureDeadline := now.AddDate(0, 0, 5)
	activeRound := models.FeedbackRound{
		SubjectID:   bobID,
		CreatedByID: adminID,
		Deadline:    &futureDeadline,
		Status:      models.RoundActive,
		CreatedAt:   now.AddDate(0, 0, -3),
		UpdatedAt:   now.AddDate(0, 0, -3),
	}
	activeResult, err := db.Collection("feedback_rounds").InsertOne(ctx, activeRound)
	if err != nil {
		log.Printf("❌ Failed to create active round: %v", err)
	} else {
		activeRoundID := activeResult.InsertedID.(primitive.ObjectID)
		log.Printf("✅ Created ACTIVE round for Bob (ID: %s)", activeRoundID.Hex())

		// Add reviewers
		reviewers := []primitive.ObjectID{aliceID, carolID, davidID}
		for _, reviewerID := range reviewers {
			reviewer := models.RoundReviewer{
				RoundID:    activeRoundID,
				ReviewerID: reviewerID,
				CreatedAt:  now.AddDate(0, 0, -3),
			}
			_, err := db.Collection("feedback_rounds").UpdateOne(
				ctx,
				bson.M{"_id": activeRoundID},
				bson.M{"$push": bson.M{"reviewers": reviewer}},
			)
			if err != nil {
				log.Printf("❌ Failed to add reviewer: %v", err)
			}
		}

		// Add one submission (Alice submitted, Carol and David haven't yet)
		aliceSubmission := models.Submission{
			RoundID:    activeRoundID,
			ReviewerID: aliceID,
			Responses: toJSON(map[string]string{
				"a": "Bob is an excellent communicator and always keeps the team informed. His technical skills are top-notch.",
				"b": "Could improve on meeting deadlines and time management. Sometimes takes on too much at once.",
				"c": "Always willing to help teammates and shares knowledge freely. Very collaborative.",
				"d": "Consider using project management tools more effectively and learning to say no to avoid overcommitment.",
			}),
			SubmittedAt: now.AddDate(0, 0, -2),
			UpdatedAt:   now.AddDate(0, 0, -2),
		}
		_, err = db.Collection("submissions").InsertOne(ctx, aliceSubmission)
		if err != nil {
			log.Printf("❌ Failed to create submission: %v", err)
		} else {
			log.Printf("✅ Created 1 submission for active round (2 pending)")
		}
	}

	// 3. CLOSED ROUND - Deadline passed, all submissions complete, ready for AI consolidation
	pastDeadline := now.AddDate(0, 0, -2)
	closedRound := models.FeedbackRound{
		SubjectID:   carolID,
		CreatedByID: adminID,
		Deadline:    &pastDeadline,
		Status:      models.RoundClosed,
		CreatedAt:   now.AddDate(0, 0, -10),
		UpdatedAt:   now.AddDate(0, 0, -2),
	}
	closedResult, err := db.Collection("feedback_rounds").InsertOne(ctx, closedRound)
	if err != nil {
		log.Printf("❌ Failed to create closed round: %v", err)
	} else {
		closedRoundID := closedResult.InsertedID.(primitive.ObjectID)
		log.Printf("✅ Created CLOSED round for Carol (ID: %s) - READY FOR CONSOLIDATION!", closedRoundID.Hex())

		// Add reviewers
		reviewers := []primitive.ObjectID{aliceID, bobID, davidID, eveID}
		for _, reviewerID := range reviewers {
			reviewer := models.RoundReviewer{
				RoundID:    closedRoundID,
				ReviewerID: reviewerID,
				CreatedAt:  now.AddDate(0, 0, -10),
			}
			_, err := db.Collection("feedback_rounds").UpdateOne(
				ctx,
				bson.M{"_id": closedRoundID},
				bson.M{"$push": bson.M{"reviewers": reviewer}},
			)
			if err != nil {
				log.Printf("❌ Failed to add reviewer: %v", err)
			}
		}

		// Create realistic feedback submissions for AI consolidation testing
		submissions := []models.Submission{
			{
				RoundID:    closedRoundID,
				ReviewerID: aliceID,
				Responses: toJSON(map[string]string{
					"a": "Carol has exceptional problem-solving skills and always finds creative solutions. She's detail-oriented and catches issues before they become problems. Her code reviews are thorough and educational.",
					"b": "Could work on public speaking and presenting to larger groups. Sometimes hesitates to voice opinions in meetings. Would benefit from more confidence in sharing ideas with senior leadership.",
					"c": "Consistently delivers high-quality work on time. Takes initiative to improve processes. Mentors junior team members effectively and patiently.",
					"d": "Consider taking on more visible projects to showcase your talents. Practice presenting technical concepts to non-technical audiences. Your expertise deserves wider recognition in the organization.",
				}),
				SubmittedAt: now.AddDate(0, 0, -8),
				UpdatedAt:   now.AddDate(0, 0, -8),
			},
			{
				RoundID:    closedRoundID,
				ReviewerID: bobID,
				Responses: toJSON(map[string]string{
					"a": "Carol's technical expertise is outstanding. She writes clean, maintainable code and sets a great example for the team. Her documentation is always comprehensive. Great at breaking down complex problems.",
					"b": "Sometimes gets too deep into technical details and loses sight of business priorities. Could benefit from more strategic thinking and understanding the 'why' behind projects. Work-life balance could be improved.",
					"c": "Very reliable and dependable. Always meets commitments. Stays late to help others when needed. Actively participates in code reviews and provides constructive feedback.",
					"d": "Focus more on the business impact of your work. Build stronger relationships with product managers. Take time to recharge - you're most effective when well-rested. Consider delegating more to grow the team.",
				}),
				SubmittedAt: now.AddDate(0, 0, -6),
				UpdatedAt:   now.AddDate(0, 0, -6),
			},
			{
				RoundID:    closedRoundID,
				ReviewerID: davidID,
				Responses: toJSON(map[string]string{
					"a": "Carol is incredibly knowledgeable about our tech stack and system architecture. She's patient when explaining complex concepts and makes time for everyone. Her analytical skills are impressive.",
					"b": "Could be more proactive in sharing knowledge through documentation or tech talks. Sometimes works in isolation and misses opportunities for collaboration. May benefit from pair programming more often.",
					"c": "Consistently produces high-quality technical solutions. Responds quickly to questions and unblocks teammates. Takes ownership of her work and follows through completely.",
					"d": "Start a weekly tech talk series to share your knowledge with the broader team. Engage more in cross-team collaborations. Your insights would be valuable in architecture discussions.",
				}),
				SubmittedAt: now.AddDate(0, 0, -4),
				UpdatedAt:   now.AddDate(0, 0, -4),
			},
			{
				RoundID:    closedRoundID,
				ReviewerID: eveID,
				Responses: toJSON(map[string]string{
					"a": "Carol has a strong attention to detail and produces work with very few bugs. She's great at debugging complex issues and helping others troubleshoot. Her testing approach is thorough and systematic.",
					"b": "Could improve communication about progress and blockers. Sometimes assumes others understand technical context that needs explanation. Could be more vocal about accomplishments during team meetings.",
					"c": "Takes time to write comprehensive tests. Always willing to review code and provide thoughtful feedback. Maintains high standards for code quality across the team.",
					"d": "Schedule regular sync-ups with stakeholders to keep them informed. Practice explaining technical work in business terms. Don't be shy about celebrating wins - your contributions are significant.",
				}),
				SubmittedAt: now.AddDate(0, 0, -3),
				UpdatedAt:   now.AddDate(0, 0, -3),
			},
		}

		for _, submission := range submissions {
			_, err := db.Collection("submissions").InsertOne(ctx, submission)
			if err != nil {
				log.Printf("❌ Failed to create submission: %v", err)
			}
		}
		log.Printf("✅ Created 4 realistic submissions for closed round")
		log.Printf("🎯 TEST THIS: Call POST /api/rounds/%s/consolidate to test AI!", closedRoundID.Hex())
	}

	// 4. SHARED ROUND - Already consolidated and shared with subject
	sharedDeadline := now.AddDate(0, 0, -15)
	sharedRound := models.FeedbackRound{
		SubjectID:   davidID,
		CreatedByID: adminID,
		Deadline:    &sharedDeadline,
		Status:      models.RoundShared,
		CreatedAt:   now.AddDate(0, 0, -20),
		UpdatedAt:   now.AddDate(0, 0, -5),
	}
	sharedResult, err := db.Collection("feedback_rounds").InsertOne(ctx, sharedRound)
	if err != nil {
		log.Printf("❌ Failed to create shared round: %v", err)
	} else {
		sharedRoundID := sharedResult.InsertedID.(primitive.ObjectID)
		log.Printf("✅ Created SHARED round for David")

		// Add reviewers
		reviewers := []primitive.ObjectID{aliceID, bobID, carolID}
		for _, reviewerID := range reviewers {
			reviewer := models.RoundReviewer{
				RoundID:    sharedRoundID,
				ReviewerID: reviewerID,
				CreatedAt:  now.AddDate(0, 0, -20),
			}
			_, err := db.Collection("feedback_rounds").UpdateOne(
				ctx,
				bson.M{"_id": sharedRoundID},
				bson.M{"$push": bson.M{"reviewers": reviewer}},
			)
			if err != nil {
				log.Printf("❌ Failed to add reviewer: %v", err)
			}
		}

		// Create submissions
		sharedSubmissions := []models.Submission{
			{
				RoundID:    sharedRoundID,
				ReviewerID: aliceID,
				Responses: toJSON(map[string]string{
					"a": "David brings great energy to the team and is always positive. Good technical foundation and eager to learn.",
					"b": "Needs to improve time estimation and communication about delays. Sometimes rushes through testing.",
					"c": "Asks good questions and seeks feedback. Participates actively in team discussions.",
					"d": "Take time to plan before coding. Practice breaking down large tasks. Reach out earlier when blocked.",
				}),
				SubmittedAt: now.AddDate(0, 0, -18),
				UpdatedAt:   now.AddDate(0, 0, -18),
			},
			{
				RoundID:    sharedRoundID,
				ReviewerID: bobID,
				Responses: toJSON(map[string]string{
					"a": "Enthusiastic and motivated. Quick to pick up new technologies. Good team player.",
					"b": "Could benefit from more systematic debugging approach. Sometimes jumps to solutions without fully understanding the problem.",
					"c": "Shows initiative and willingness to tackle challenging tasks. Responsive to feedback.",
					"d": "Focus on fundamentals before adding complexity. Build more robust tests. Document your learning process.",
				}),
				SubmittedAt: now.AddDate(0, 0, -17),
				UpdatedAt:   now.AddDate(0, 0, -17),
			},
			{
				RoundID:    sharedRoundID,
				ReviewerID: carolID,
				Responses: toJSON(map[string]string{
					"a": "Great attitude and willingness to help others. Creative problem solver. Brings fresh perspectives.",
					"b": "Needs to improve code organization and follow established patterns more consistently. Testing practices need work.",
					"c": "Actively seeks learning opportunities. Accepts feedback well and implements suggestions.",
					"d": "Study the codebase more thoroughly before making changes. Write tests first. Pair program with senior engineers.",
				}),
				SubmittedAt: now.AddDate(0, 0, -16),
				UpdatedAt:   now.AddDate(0, 0, -16),
			},
		}

		for _, submission := range sharedSubmissions {
			_, err := db.Collection("submissions").InsertOne(ctx, submission)
			if err != nil {
				log.Printf("❌ Failed to create submission: %v", err)
			}
		}

		// Create a consolidation for this round
		sharedAt := now.AddDate(0, 0, -5)
		consolidation := models.Consolidation{
			RoundID:          sharedRoundID,
			GeneratedByID:    adminID,
			ExecutiveSummary: "David demonstrates strong enthusiasm and positive energy that benefits the team. While he shows good initiative and learning ability, there are opportunities to improve technical practices, particularly around planning, testing, and systematic problem-solving. With more focus on fundamentals and structured development practices, David can accelerate his growth trajectory.",
			Strengths: toJSON([]string{
				"Positive attitude and team energy",
				"Quick learner and eager to adopt new technologies",
				"Good team player who actively participates",
				"Creative problem-solving approach",
				"Responsive to feedback and implements suggestions",
			}),
			AreasForImprovement: toJSON([]string{
				"Time estimation and communication about delays",
				"Testing practices and thoroughness",
				"Systematic debugging approach",
				"Code organization and following established patterns",
				"Understanding problems fully before jumping to solutions",
			}),
			ActionableInsights: toJSON([]string{
				"Practice breaking down large tasks into smaller, estimable pieces",
				"Develop a systematic debugging checklist and follow it consistently",
				"Write tests before implementing features (TDD approach)",
				"Study the codebase architecture before making changes",
				"Schedule regular pair programming sessions with senior engineers",
				"Reach out earlier when blocked rather than struggling alone",
			}),
			QuestionSummaries: toJSON(map[string]string{
				"a": "Team members consistently highlighted David's enthusiasm, positive energy, and eagerness to learn. His willingness to help others and participate actively in team discussions was noted by multiple reviewers.",
				"b": "Key areas for improvement include time estimation, testing practices, and taking a more systematic approach to debugging. Reviewers noted a tendency to rush and jump to solutions without fully understanding problems.",
				"c": "David demonstrates initiative, asks good questions, and accepts feedback well. He shows creativity in problem-solving and actively seeks learning opportunities.",
				"d": "Reviewers recommend focusing on fundamentals, planning before coding, writing tests first, and engaging more in pair programming with senior team members.",
			}),
			AdminNotes: "David has shown good progress this quarter. Recommend assigning a dedicated mentor for the next 3 months to help build stronger foundations. Consider enrolling in the company's advanced testing workshop.",
			SharedAt:   &sharedAt,
			CreatedAt:  now.AddDate(0, 0, -6),
			UpdatedAt:  now.AddDate(0, 0, -5),
		}

		_, err = db.Collection("consolidations").InsertOne(ctx, consolidation)
		if err != nil {
			log.Printf("❌ Failed to create consolidation: %v", err)
		} else {
			log.Printf("✅ Created and shared consolidation for David")
		}
	}

	log.Println("\n✨ Development data seeding complete!")
	log.Println("\n📋 Summary:")
	log.Println("  • 6 users (1 admin, 5 members)")
	log.Println("  • 4 feedback rounds in different statuses:")
	log.Println("    - DRAFT: Alice (no reviewers yet)")
	log.Println("    - ACTIVE: Bob (3 reviewers, 1 submission)")
	if closedResult != nil {
		log.Printf("    - CLOSED: Carol (4 reviewers, 4 submissions) → ID: %s\n", closedResult.InsertedID.(primitive.ObjectID).Hex())
		log.Printf("      🎯 Ready to test AI consolidation!\n")
		log.Printf("      Run: POST /api/rounds/%s/consolidate\n", closedResult.InsertedID.(primitive.ObjectID).Hex())
	}
	log.Println("    - SHARED: David (3 reviewers, 3 submissions, consolidation shared)")
	log.Println("\n🔑 Login credentials:")
	log.Println("  • admin@example.com (Admin)")
	log.Println("  • alice@example.com, bob@example.com, carol@example.com, etc. (Members)")
	log.Println("\n💡 Use /api/auth/dev-login?email=admin@example.com to login in dev mode")
}

func toJSON(v interface{}) string {
	bytes, _ := json.Marshal(v)
	return string(bytes)
}
