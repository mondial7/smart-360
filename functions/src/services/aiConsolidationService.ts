/**
 * AI Consolidation Service
 *
 * Uses OpenAI to consolidate anonymous feedback into actionable insights
 */

import { openai } from '../config/openai-client';
import * as functions from 'firebase-functions';

export interface FeedbackAnswers {
  q1: string;
  q2: string;
  q3: string;
  q4: string;
}

export interface FeedbackSubmission {
  id: string;
  roundId: string;
  reviewerId: string;
  subjectId: string;
  answers: FeedbackAnswers;
  submittedAt: Date;
  isAnonymous: boolean;
}

export interface AISummary {
  overview: string;
  keyStrengths: string[];
  areasForImprovement: string[];
  actionableInsights: string[];
  q1Summary: string;
  q2Summary: string;
  q3Summary: string;
  q4Summary: string;
}

/**
 * Anonymize feedback submissions by removing reviewer identifiers
 */
const anonymizeFeedback = (submissions: FeedbackSubmission[]): Array<{
  reviewer: string;
  answers: FeedbackAnswers;
}> => {
  return submissions.map((submission, index) => ({
    reviewer: `Reviewer ${index + 1}`,
    answers: submission.answers,
  }));
};

/**
 * Generate the prompt for OpenAI to consolidate feedback
 */
const generateConsolidationPrompt = (
  anonymizedFeedback: Array<{ reviewer: string; answers: FeedbackAnswers }>,
  questions: {
    q1: string;
    q2: string;
    q3: string;
    q4: string;
  }
): string => {
  const feedbackText = anonymizedFeedback.map((feedback) => `
${feedback.reviewer}:
- ${questions.q1}: ${feedback.answers.q1}
- ${questions.q2}: ${feedback.answers.q2}
- ${questions.q3}: ${feedback.answers.q3}
- ${questions.q4}: ${feedback.answers.q4}
`).join('\n');

  return `You are a professional HR consultant analyzing 360-degree feedback for a colleague.
You have ${anonymizedFeedback.length} anonymous feedback responses.

Your task:
1. Synthesize feedback while maintaining complete anonymity (NEVER reference specific reviewers)
2. Extract key themes and patterns across all responses
3. Identify strengths and areas for improvement
4. Provide actionable, specific insights for professional development

The 4 feedback questions are:
1. ${questions.q1}
2. ${questions.q2}
3. ${questions.q3}
4. ${questions.q4}

Anonymous Feedback Responses:
${feedbackText}

Please provide a JSON response with the following structure:
{
  "overview": "A 2-3 sentence executive summary of the overall feedback themes",
  "keyStrengths": ["strength1", "strength2", "strength3"],
  "areasForImprovement": ["area1", "area2", "area3"],
  "actionableInsights": ["specific actionable insight 1", "specific actionable insight 2", "specific actionable insight 3"],
  "q1Summary": "Synthesized summary of all responses to question 1",
  "q2Summary": "Synthesized summary of all responses to question 2",
  "q3Summary": "Synthesized summary of all responses to question 3",
  "q4Summary": "Synthesized summary of all responses to question 4"
}

IMPORTANT:
- Be honest and constructive
- Identify patterns and themes across multiple responses
- Be specific and actionable
- Never reveal which reviewer said what
- If responses conflict, acknowledge different perspectives
- Use professional, supportive language`;
};

/**
 * Consolidate feedback using OpenAI
 */
export const consolidateFeedbackWithAI = async (
  submissions: FeedbackSubmission[],
  questions: {
    q1: string;
    q2: string;
    q3: string;
    q4: string;
  }
): Promise<AISummary> => {
  try {
    // Validate minimum submissions
    if (submissions.length < 2) {
      throw new Error('Need at least 2 feedback submissions for consolidation');
    }

    // Anonymize feedback
    const anonymizedFeedback = anonymizeFeedback(submissions);

    // Generate prompt
    const prompt = generateConsolidationPrompt(anonymizedFeedback, questions);

    functions.logger.info('Calling OpenAI API for feedback consolidation', {
      submissionCount: submissions.length,
    });

    // Call OpenAI API
    const response = await openai.chat.completions.create({
      model: 'gpt-4o-mini',
      messages: [
        {
          role: 'system',
          content: 'You are a professional HR consultant specializing in 360-degree feedback analysis. You provide honest, constructive, and actionable insights.',
        },
        {
          role: 'user',
          content: prompt,
        },
      ],
      response_format: { type: 'json_object' },
      temperature: 0.7,
      max_tokens: 2000,
    });

    const aiResponse = response.choices[0].message.content;

    if (!aiResponse) {
      throw new Error('OpenAI returned empty response');
    }

    // Parse JSON response
    const parsedResponse = JSON.parse(aiResponse) as AISummary;

    functions.logger.info('Successfully consolidated feedback with AI', {
      submissionCount: submissions.length,
      tokensUsed: response.usage?.total_tokens,
    });

    return parsedResponse;
  } catch (error: any) {
    functions.logger.error('Error consolidating feedback with AI', {
      error: error.message,
      stack: error.stack,
    });

    throw new functions.https.HttpsError(
      'internal',
      `Failed to consolidate feedback: ${error.message}`
    );
  }
};
