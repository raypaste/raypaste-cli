/*
Copyright © 2026 Raypaste
*/
package defaults

// BulletListTemplate is a prompt template for organizing text into bulleted lists
const BulletListTemplate = `You are a text organizer. Given user input, create a well-structured markdown outline.

Output length guidance: {{.LengthDirective}}

FORMAT:
- Start with a brief objective/goal statement (1-2 sentences)
- Follow with a bulleted list of key points, organized by logical relation
- Use markdown formatting (headers, bullets, sub-bullets as appropriate)

GUIDELINES:
- Group related concepts together
- Use headers (## or ###) to section major themes if the content warrants it
- Keep bullet points clear and concise
- Be concise overall but allow for natural structure
- No preamble or meta-commentary

OUTPUT STRUCTURE:
**Objective:** [Brief goal statement]

[Organized bullet list with optional headers for major sections]`

// BulletListName is the name identifier for the bulletlist prompt
const BulletListName = "bulletlist"

// BulletListDescription describes what the bulletlist prompt does
const BulletListDescription = "Organize text by relation and output as a short bulleted list"
