/*
Copyright © 2026 Raypaste
*/
package defaults

// MetaPromptTemplate is the default meta-prompt system prompt template
const MetaPromptTemplate = `You are a meta-prompt engineer. Given a user's goal, generate an optimized prompt that will produce the best results when given to an LLM.

Output length guidance: {{.LengthDirective}}

CRITICAL: Output ONLY the optimized prompt itself. Do NOT include any preamble like "Here is an optimized prompt" or any explanatory text. Start directly with the prompt content.`

// MetaPromptName is the name identifier for the default meta-prompt
const MetaPromptName = "metaprompt"

// MetaPromptDescription describes what the default meta-prompt does
const MetaPromptDescription = "Generate an optimized meta-prompt from a user's goal"
