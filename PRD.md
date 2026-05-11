# Smart 360 Feedback - Product Requirements Document

## Product Overview

Smart 360 Feedback is a web application designed to facilitate anonymous, structured peer feedback within organizations. The platform enables administrators to coordinate feedback collection cycles, leverages artificial intelligence to synthesize feedback into actionable insights, and delivers consolidated results to team members for professional development.

## Problem Statement

Traditional 360-degree feedback processes are often:
- Time-consuming and difficult to coordinate
- Challenging to maintain reviewer anonymity
- Hard to synthesize into actionable insights
- Cumbersome for both administrators and participants

## Goals

1. **Streamline Feedback Collection**: Simplify the process of requesting and collecting peer feedback
2. **Ensure Anonymity**: Protect reviewer identity throughout the entire process
3. **Generate Actionable Insights**: Transform raw feedback into structured, actionable development recommendations
4. **Save Time**: Reduce administrative overhead through automation and AI consolidation
5. **Improve Engagement**: Make feedback submission and consumption straightforward and accessible

## User Roles

### Administrator
Organization leaders or HR professionals who:
- Manage team member roster and team definitions
- Initiate and coordinate feedback rounds
- Review and consolidate feedback
- Share results with feedback subjects

### Team Administrator
Team leads who:
- Manage members within their team
- Create and coordinate feedback rounds for their team
- View consolidated feedback for their team

### Team Member
Individual contributors who:
- Provide feedback on peers when requested
- Receive consolidated feedback for their own development (in-app and as a downloadable PDF)
- Track pending feedback requests
- Track personal feedback insights through analytics on the dashboard

## Core Features

### 1. Authentication & Access Control

**Capability**: Secure access management for all users

**Key Functions**:
- Single sign-on using Google accounts
- First user automatically assigned administrator role
- Role-based access control (Administrator vs. Team Member)
- User profile display (name, email, photo)

### 2. Team Management

**Capability**: Maintain organization roster and role assignments

**Key Functions**:
- View all team members in a centralized directory
- Display member information: name, email, role, join date, last login
- Assign and modify user roles (admin/member)
- Track team member status and activity

**Access Control**:
- Administrators: Full read/write access
- Team Members: Read-only view of team roster

### 3. Feedback Round Creation

**Capability**: Initiate structured feedback collection cycles

**Multi-Step Wizard**:
1. Select the feedback subject (person receiving feedback)
2. Choose multiple reviewers from team roster
3. Set submission deadline
4. Review and confirm round details

**Round Configuration**:
- Exclude subject from their own review
- Exclude administrator creating the round
- Set clear submission deadlines
- Standard question set across all rounds

**Standard Feedback Questions**:
1. What are this person's key strengths?
2. What areas could this person improve?
3. What specific behaviors or actions have you observed that stood out?
4. What advice would you give to help this person grow?

**Round Statuses**:
- **Draft**: Round created but not finalized
- **Active**: Round in progress, awaiting submissions
- **Closed**: Submission deadline passed
- **Shared**: Consolidated feedback delivered to subject

**Access Control**: Administrator only

### 4. Feedback Submission

**Capability**: Anonymous peer feedback collection

**Submission Process**:
- Reviewers see list of pending feedback requests
- Access dedicated submission form for each round
- Respond to four structured questions
- View submission deadline
- Submit anonymously (one submission per reviewer per round)

**Submission Features**:
- Text-based responses to structured questions
- Real-time status tracking (pending vs. submitted)
- Deadline visibility and validation
- Confirmation of successful submission

**Anonymity Protection**:
- Reviewer identity never revealed to subject
- Administrator cannot see individual reviewer responses
- Submissions aggregated before consolidation

**Access Control**: All team members (assigned reviewers)

### 5. AI-Powered Feedback Consolidation

**Capability**: Automated synthesis of raw feedback into structured insights

**Consolidation Process**:
1. Administrator triggers consolidation after round closes
2. All anonymous submissions aggregated
3. AI analyzes feedback and generates comprehensive summary
4. Structured output organized by theme

**Generated Output Sections**:
- **Executive Summary**: High-level overview of feedback themes
- **Key Strengths** (3+): Identified positive attributes and behaviors
- **Areas for Improvement** (3+): Development opportunities
- **Actionable Insights** (3+): Specific, concrete recommendations
- **Question-by-Question Summary**: Detailed synthesis for each of the four questions

**Quality Assurance**:
- AI instructed to maintain anonymity in synthesis
- Generic references only (no reviewer attribution)
- Balanced, constructive tone
- Focus on patterns across multiple submissions

**Access Control**: Administrator only

### 6. Feedback Customization & Delivery

**Capability**: Personalized delivery of consolidated feedback to subjects

**Admin Customization**:
- Review AI-generated consolidation
- Add optional admin notes for context or guidance
- Finalize and share with subject

**Subject Access**:
- View all shared feedback rounds in centralized location
- Access detailed consolidated insights
- See number of participating reviewers
- Read admin notes (if provided)
- Identify AI-synthesized content

**Access Control**:
- Sharing: Administrator only
- Viewing: Subject and administrators

### 7. PDF Export

**Capability**: Download a branded, print-ready PDF of any consolidated feedback

**Behavior**:
- Available from the "My Feedback" view on each shared consolidation card
- Endpoint: `GET /api/consolidations/:roundId/pdf`
- Authorization rules:
  - Admins: any consolidation
  - Round subject: their own consolidation, only after it has been shared
  - Round creator: consolidations for rounds they created
- Filename pattern: `smart360-<subject-slug>-<round-date>.pdf`

**Content sections** (included when populated):
- Header with subject, round date, and shared date
- Executive summary
- Key strengths (bulleted)
- Areas for improvement (bulleted)
- Actionable insights (bulleted)
- Question-by-question summary (4 sections)
- Admin notes (if present)
- Confidentiality footer

**Access Control**: Subject (after sharing), round creator, admin

### 8. Personal Analytics

**Capability**: Self-service insights on the user dashboard

**Endpoint**: `GET /api/analytics/me`

**Returned data**:
- Number of consolidations shared with the user
- Number of submissions the user has made as a reviewer
- Pending review count
- Per-round breakdown: round id, shared date, strengths count, improvements count, insights count, presence of behaviors/growth summaries
- Latest-round radar axes: strengths, improvements, behaviors, growth

**Visualization**:
- Three at-a-glance counters (received / submitted / pending)
- SVG radar chart of the latest shared round (4 axes)
- Bar trend (strengths / improvements / insights) per shared round, ordered chronologically

**Access Control**: Each user sees only their own analytics

### 9. Real-Time Dashboards

**Capability**: At-a-glance overview of system status and pending actions

**Administrator Dashboard**:
- Total users count
- Admin vs. member breakdown
- Total feedback rounds count
- List of active rounds
- Quick navigation to team management and round creation

**Team Member Dashboard**:
- Pending feedback requests (with deadlines)
- Notification when new feedback shared
- Direct access to "My Feedback" view
- Submission status for each request

**Real-Time Updates**:
- Dashboard refreshes automatically as data changes
- Live submission count updates
- Status changes reflected immediately

## User Workflows

### Primary Workflow: Complete Feedback Cycle

1. **Administrator creates feedback round**
   - Selects subject needing feedback
   - Chooses 3-5 peer reviewers
   - Sets deadline (e.g., 2 weeks)
   - Confirms and activates round

2. **Reviewers see assignment in dashboard**
   - Pending feedback requests appear on the dashboard with deadline
   - Direct link to submission form

3. **Reviewers submit feedback**
   - Access pending feedback from dashboard
   - Complete four structured questions
   - Submit anonymously before deadline
   - Receive confirmation

4. **Round closes at deadline**
   - Status changes to "closed"
   - No further submissions accepted
   - Ready for consolidation

5. **Administrator consolidates feedback**
   - Reviews submission count (not individual responses)
   - Triggers AI consolidation
   - Reviews generated insights
   - Adds optional admin notes

6. **Administrator shares with subject**
   - Finalizes consolidated feedback
   - Marks as "shared"
   - Subject sees the new consolidation in their "My Feedback" view

7. **Subject reviews feedback**
   - Accesses "My Feedback" section
   - Reads consolidated insights
   - Optionally downloads as PDF
   - Reviews dashboard analytics (radar + trends across rounds)
   - Uses insights for development planning

### Secondary Workflow: Team Management

1. **New team member joins organization**
   - Signs in with Google account
   - Profile automatically created
   - Default role: Team Member
   - Appears in team roster

2. **Administrator assigns roles**
   - Views team member list
   - Promotes members to admin as needed
   - Manages role assignments

### Tertiary Workflow: Viewing Historical Feedback

1. **Team member checks development history**
   - Navigates to "My Feedback"
   - Views all shared feedback rounds
   - Reviews past consolidated insights
   - Tracks progress over time

## Data & Privacy

### Anonymity Guarantee
- Reviewer identity never stored with submission
- Administrator cannot view individual reviewer responses
- AI consolidation uses generic labels (Reviewer 1, Reviewer 2)
- Only aggregated, anonymized feedback shared with subject

### Data Access Rules
- **Administrators**: Can read all rounds, consolidated feedback, team roster
- **Team Members**: Can only read their own submissions and shared feedback
- **Subjects**: Can only view feedback explicitly shared with them
- **Raw Submissions**: Only accessible to system (not visible to any user)

### Audit Trail
- All actions tracked with timestamps
- Created by, consolidated by, shared by metadata
- Last login tracking for team members
- Submission timestamps recorded

## Success Metrics

### Engagement Metrics
- Percentage of reviewers completing feedback before deadline
- Average time from request to submission
- Number of active feedback rounds per month
- Subject engagement with shared feedback

### Quality Metrics
- Average number of reviewers per round
- Consolidation turnaround time (close to share)
- Admin notes addition rate

### System Health
- Email delivery success rate
- Reminder effectiveness (submission rate after reminder)
- User login frequency
- Active user percentage

## Production Readiness

### Scalability Considerations
- Designed for teams of 10-500 members
- Supports concurrent feedback rounds
- Automated reminders scale with team size
- AI consolidation optimized for cost efficiency

### Cost Structure
Estimated monthly operational cost for small team (20 users, 10 rounds/month):
- Database operations: ~$1-2
- AI consolidation: ~$0.05
- Hosting and email: Minimal
- **Total: ~$1-2/month**

### Monitoring & Support
- System logs for troubleshooting
- Email delivery tracking
- AI consolidation success monitoring
- User activity analytics

## Future Enhancements (Not Currently Implemented)

Potential features for future releases — tracked as [open issues](https://github.com/mondial7/smart-360/issues):
- Admin-level analytics dashboard (org-wide trends, themes, completion rates) — _shipped, see `PRODUCT-OVERVIEW.md`_
- [360 comparison across multiple rounds](https://github.com/mondial7/smart-360/issues/29)
- [Anonymous comments on shared consolidated feedback](https://github.com/mondial7/smart-360/issues/31)
- [Slack / Teams integrations](https://github.com/mondial7/smart-360/issues/30)
- [Self-nomination for feedback](https://github.com/mondial7/smart-360/issues/41)
- [Peer recognition system](https://github.com/mondial7/smart-360/issues/42)
- [Manager-specific workflows](https://github.com/mondial7/smart-360/issues/43)
