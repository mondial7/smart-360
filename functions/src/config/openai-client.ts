/**
 * OpenAI Client Configuration
 */

import OpenAI from 'openai';
import * as functions from 'firebase-functions';

// Initialize OpenAI client with API key from Firebase config
const getOpenAIClient = (): OpenAI => {
  const apiKey = functions.config().openai?.api_key || process.env.OPENAI_API_KEY;

  if (!apiKey) {
    throw new Error('OpenAI API key not configured. Run: firebase functions:config:set openai.api_key="your-key"');
  }

  return new OpenAI({
    apiKey,
  });
};

export const openai = getOpenAIClient();
