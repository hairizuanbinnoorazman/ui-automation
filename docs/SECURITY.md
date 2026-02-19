# Security - Prompt Injection Mitigation

This document describes the security measures implemented to protect the script generation feature from prompt injection attacks.

## Overview

The script generation feature (`POST /api/v1/procedures/{procedure_id}/scripts`) uses AWS Bedrock with Claude to generate Python automation scripts from test procedure content. User-provided fields (Name, Description, Steps) are embedded into LLM prompts, creating potential prompt injection vulnerabilities where malicious users could craft procedure content to manipulate the LLM's behavior.

## Multi-Layer Defense Strategy

We implement defense-in-depth with three complementary layers:

### 1. Input Validation Layer

**Purpose:** Enforce strict length limits and content rules when test procedures are used for script generation.

**Length Limits (configurable via `config.yaml`):**
- Name: 255 characters (matches DB schema)
- Description: 5,000 characters
- Steps JSON: 50,000 characters serialized
- Total steps count: 200 steps maximum

**Content Validation:**
- Name: Alphanumeric + spaces, hyphens, underscores, parentheses only
- Description: Block excessive control characters and non-printable characters
- Steps: Validate known action types (`navigate`, `click`, `type`, `wait`, `assert_text`, `screenshot`) and required field names

**Implementation:**
- `testprocedure/validator.go`: `ValidateForScriptGeneration()` function
- Called before building prompts in `scriptgen/prompt.go:BuildPrompt()`

### 2. Content Sanitization Layer

**Purpose:** Sanitize user content before embedding in prompts while preserving legitimate use cases.

**Sanitization Operations:**
- Remove control characters (except `\n`, `\t`, `\r` where appropriate)
- Normalize whitespace (collapse multiple spaces/newlines)
- Strip non-printable characters
- Validate and sanitize URLs in Steps (add `https://` prefix if missing)
- Sanitize CSS selectors to remove control characters

**Implementation:**
- `scriptgen/sanitizer.go`:
  - `SanitizeTestProcedureName()` - Name sanitization
  - `SanitizeTestProcedureDescription()` - Description sanitization
  - `SanitizeSteps()` - Steps validation and sanitization

### 3. Prompt Structure Hardening

**Purpose:** Use XML-style delimiters to create clear boundaries between instructions and user data, following Anthropic's prompt engineering best practices.

**Hardened Structure:**
```
Generate a Python automation script for the following test procedure.

<test_procedure>
<name>{sanitized_name}</name>
<version>{version}</version>
<description>{sanitized_description}</description>
<test_steps>
{sanitized_steps_json}
</test_steps>
</test_procedure>

<requirements>
[System instructions...]
</requirements>
```

**Why XML Tags?**
- Creates clear boundaries between different sections of the prompt
- Makes it harder to "break out" of the user data section
- Claude models are trained to respect these structural boundaries
- Standard practice recommended by Anthropic for prompt security

**Implementation:**
- `scriptgen/prompt.go:BuildPrompt()` - Uses XML structure for prompt formatting

## Suspicious Pattern Detection

The system detects and blocks common prompt injection patterns:

- "ignore previous instructions"
- "ignore all previous"
- "disregard previous"
- "forget all previous"
- "new instructions:"
- "system:"
- XML tag breakout attempts (`</test_procedure>`, `<requirements>`, etc.)
- Excessive control characters (>5% of content)

**Implementation:**
- `testprocedure/validator.go:checkSuspiciousPatterns()` - Case-insensitive detection

## Configuration

Validation limits are configurable in `config.yaml`:

```yaml
script_gen:
  validation:
    max_name_length: 255           # Maximum length for procedure name
    max_description_length: 5000   # Maximum length for procedure description
    max_steps_json_length: 50000   # Maximum length for serialized steps JSON
    max_steps_count: 200            # Maximum number of steps in a procedure
  monitoring:
    log_suspicious_patterns: true  # Log when suspicious patterns are detected
```

Environment variables can override config file settings using the pattern `SCRIPT_GEN_VALIDATION_MAX_NAME_LENGTH=500`.

## Backward Compatibility

**Existing Procedures:**
Existing test procedures in the database may not meet new validation rules. Strategy:
- Validation is applied only during script generation (not on procedure create/update)
- If an existing procedure fails validation, an error is returned with clear messaging
- No database migration needed
- Users can update their procedures to meet validation requirements

## Error Handling

**Clear Error Messages:**
When validation fails, the API returns descriptive error messages:

- `"name exceeds maximum length of 255 characters"`
- `"description contains suspicious pattern 'ignore previous instructions'"`
- `"step 3: invalid action type 'delete'"`
- `"step 5 (navigate) missing required 'url' field"`

**HTTP Status Codes:**
- `400 Bad Request` - Validation failures, invalid content
- `500 Internal Server Error` - System errors during generation

## Testing Strategy

**Unit Tests:**
- `scriptgen/sanitizer_test.go` - Tests for all sanitization functions
- `testprocedure/validator_test.go` - Tests for validation logic
- `scriptgen/prompt_test.go` - Tests for prompt generation with security checks

**Test Coverage:**
- Valid content passes through unchanged
- Control characters are removed
- Length limits are enforced
- Invalid step actions are rejected
- Suspicious patterns are detected
- XML structure is correctly formed
- Injection attempts are neutralized

**Run Tests:**
```bash
make test
go test -v ./scriptgen/...
go test -v ./testprocedure/...
```

## Security Best Practices for Development

**When Adding New Features:**

1. **Always validate user input** before using it in prompts
2. **Use the existing sanitization functions** rather than creating new ones
3. **Maintain XML structure** in prompts with clear boundaries
4. **Test with malicious input** (see injection attack patterns in tests)
5. **Document any new validation rules** in this file

**Common Pitfalls to Avoid:**

- ❌ Directly interpolating user input into prompts
- ❌ Skipping validation for "trusted" users
- ❌ Removing sanitization for "backwards compatibility"
- ❌ Allowing arbitrary control characters
- ❌ Using string concatenation without XML structure

**Recommended Approach:**

- ✅ Always call `ValidateForScriptGeneration()` first
- ✅ Use sanitization functions for all user content
- ✅ Maintain XML tag boundaries in prompts
- ✅ Test with injection payloads from test suite
- ✅ Log suspicious patterns for monitoring

## Monitoring and Alerting

**Security Logging:**
The system can log security-relevant events:
- Validation failures with suspicious patterns
- Procedures exceeding warning thresholds (configurable)
- Blocked injection attempts

**Future Enhancements:**
- Metrics for validation failures by pattern type
- Alerting on repeated injection attempts from same user
- Automated blocking of users with multiple violations

## References

**Anthropic Resources:**
- [Prompt Engineering Guide](https://docs.anthropic.com/claude/docs/prompt-engineering)
- [Claude Safety Best Practices](https://docs.anthropic.com/claude/docs/claude-safety)

**Security Standards:**
- OWASP Top 10 for LLM Applications
- OWASP Input Validation Cheat Sheet

## Version History

- **2026-02-18**: Initial implementation of multi-layer defense
  - Input validation with length limits
  - Content sanitization
  - XML-structured prompts
  - Suspicious pattern detection
