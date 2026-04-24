// MongoDB initialization script
// This script runs when the container starts for the first time

// Switch to the smart360 database
db = db.getSiblingDB('smart360');

// Create collections with validation
db.createCollection('users', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['email', 'name', 'role'],
      properties: {
        email: {
          bsonType: 'string',
          description: 'must be a string and is required'
        },
        name: {
          bsonType: 'string',
          description: 'must be a string and is required'
        },
        role: {
          enum: ['admin', 'member'],
          description: 'must be either admin or member'
        },
        photo_url: {
          bsonType: 'string',
          description: 'must be a string'
        },
        last_login: {
          bsonType: 'date',
          description: 'must be a date'
        },
        created_at: {
          bsonType: 'date',
          description: 'must be a date'
        },
        updated_at: {
          bsonType: 'date',
          description: 'must be a date'
        }
      }
    }
  }
});

db.createCollection('feedback_rounds', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['subject_id', 'created_by_id', 'status'],
      properties: {
        subject_id: {
          bsonType: 'objectId',
          description: 'must be an objectId and is required'
        },
        created_by_id: {
          bsonType: 'objectId',
          description: 'must be an objectId and is required'
        },
        status: {
          enum: ['draft', 'active', 'closed', 'shared'],
          description: 'must be one of the status values'
        },
        deadline: {
          bsonType: 'date',
          description: 'must be a date'
        },
        created_at: {
          bsonType: 'date',
          description: 'must be a date'
        },
        updated_at: {
          bsonType: 'date',
          description: 'must be a date'
        }
      }
    }
  }
});

db.createCollection('round_reviewers', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['round_id', 'reviewer_id'],
      properties: {
        round_id: {
          bsonType: 'objectId',
          description: 'must be an objectId and is required'
        },
        reviewer_id: {
          bsonType: 'objectId',
          description: 'must be an objectId and is required'
        },
        created_at: {
          bsonType: 'date',
          description: 'must be a date'
        }
      }
    }
  }
});

db.createCollection('submissions', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['round_id', 'reviewer_id', 'responses', 'submitted_at'],
      properties: {
        round_id: {
          bsonType: 'objectId',
          description: 'must be an objectId and is required'
        },
        reviewer_id: {
          bsonType: 'objectId',
          description: 'must be an objectId and is required'
        },
        responses: {
          bsonType: 'string',
          description: 'must be a string (JSON) and is required'
        },
        submitted_at: {
          bsonType: 'date',
          description: 'must be a date and is required'
        }
      }
    }
  }
});

db.createCollection('consolidations', {
  validator: {
    $jsonSchema: {
      bsonType: 'object',
      required: ['round_id', 'generated_by_id'],
      properties: {
        round_id: {
          bsonType: 'objectId',
          description: 'must be an objectId and is required'
        },
        generated_by_id: {
          bsonType: 'objectId',
          description: 'must be an objectId and is required'
        },
        executive_summary: {
          bsonType: 'string',
          description: 'must be a string'
        },
        strengths: {
          bsonType: 'string',
          description: 'must be a string (JSON array)'
        },
        areas_for_improvement: {
          bsonType: 'string',
          description: 'must be a string (JSON array)'
        },
        actionable_insights: {
          bsonType: 'string',
          description: 'must be a string (JSON array)'
        },
        question_summaries: {
          bsonType: 'string',
          description: 'must be a string (JSON object)'
        },
        admin_notes: {
          bsonType: 'string',
          description: 'must be a string'
        },
        shared_at: {
          bsonType: 'date',
          description: 'must be a date'
        },
        created_at: {
          bsonType: 'date',
          description: 'must be a date'
        },
        updated_at: {
          bsonType: 'date',
          description: 'must be a date'
        }
      }
    }
  }
});

// Create indexes for better performance
print('Creating indexes...');

// Users collection indexes
db.users.createIndex({ email: 1 }, { unique: true });
db.users.createIndex({ role: 1 });
db.users.createIndex({ created_at: 1 });

// Feedback rounds collection indexes
db.feedback_rounds.createIndex({ subject_id: 1 });
db.feedback_rounds.createIndex({ created_by_id: 1 });
db.feedback_rounds.createIndex({ status: 1 });
db.feedback_rounds.createIndex({ created_at: 1 });
db.feedback_rounds.createIndex({ subject_id: 1, status: 1 });

// Round reviewers collection indexes
db.round_reviewers.createIndex({ round_id: 1 });
db.round_reviewers.createIndex({ reviewer_id: 1 });
db.round_reviewers.createIndex({ round_id: 1, reviewer_id: 1 }, { unique: true });

// Submissions collection indexes
db.submissions.createIndex({ round_id: 1 });
db.submissions.createIndex({ reviewer_id: 1 });
db.submissions.createIndex({ round_id: 1, reviewer_id: 1 }, { unique: true });
db.submissions.createIndex({ submitted_at: 1 });

// Consolidations collection indexes
db.consolidations.createIndex({ round_id: 1 }, { unique: true });
db.consolidations.createIndex({ generated_by_id: 1 });
db.consolidations.createIndex({ shared_at: 1 });
db.consolidations.createIndex({ created_at: 1 });

print('MongoDB initialization completed successfully!');
