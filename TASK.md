# ROLE: Senior Go Backend Engineer

You are now responsible for executing the technical improvements identified in:

`docs/GO_BACKEND_AUDIT.md`

The audit has already been completed. Your task is to turn the audit findings into actual code improvements.

This is an implementation task, not another audit-only task.

## 1. Initial Instructions

Before making any changes:

1. Read `docs/GO_BACKEND_AUDIT.md` completely.
2. Explore the relevant source files for each finding.
3. Verify that each finding is valid against the current implementation.
4. Understand the existing architecture, conventions, and dependencies.
5. Identify potential side effects and dependencies between fixes.
6. Check the existing tests and API contracts.

Do not blindly trust every recommendation in the audit report. Verify it against the actual code.

## 2. Implementation Priorities

Implement improvements in this order:

### Phase 1 — Critical Security Issues
Address:
- Authentication vulnerabilities.
- JWT validation and token security.
- Password hashing and verification.
- Authorization bypasses.
- Privilege escalation.
- Sensitive information exposure.
- SQL injection.
- Missing security validations.
- Other CRITICAL security findings.

### Phase 2 — High Priority Issues
Address:
- Incorrect business logic.
- Database transaction problems.
- Data consistency issues.
- Incorrect error handling.
- Race conditions.
- Resource leaks.
- Missing request validation.
- HTTP middleware problems.
- HIGH severity findings.

### Phase 3 — Architecture & Maintainability
Address:
- Poor separation of concerns.
- Unnecessary code duplication.
- Incorrect dependency direction.
- Poor package organization.
- Inconsistent naming conventions.
- Unnecessary abstractions.
- Other maintainability issues.

Preserve the existing architecture whenever possible.

### Phase 4 — Performance & Scalability
Address:
- Inefficient database queries.
- N+1 queries.
- Missing database indexes when justified.
- Unnecessary allocations.
- Inefficient data processing.
- Goroutine leaks.
- Unnecessary blocking operations.
- Caching improvements when appropriate.

Do not introduce premature optimizations.

### Phase 5 — Testing & Production Readiness
Address:
- Missing unit tests.
- Missing integration tests.
- Authentication and authorization test cases.
- Middleware tests.
- Edge cases.
- Graceful shutdown.
- HTTP server timeouts.
- Logging and observability.
- Configuration validation.
- Deployment-related issues.

## 3. Implementation Rules

Follow these rules strictly:

1. Work on the actual source code, not just documentation.
2. Fix issues based on their verified severity.
3. Make changes incrementally.
4. Keep changes focused and avoid unrelated refactoring.
5. Follow idiomatic Go practices.
6. Respect the existing Chi router implementation.
7. Preserve existing API contracts and response structures unless a change is necessary for security or correctness.
8. Do not introduce new dependencies without strong technical justification.
9. Never hardcode secrets or credentials.
10. Use context.Context appropriately for HTTP requests and database operations.
11. Ensure database connections, transactions, and other resources are handled correctly.
12. Add or update tests for every meaningful change.
13. Do not remove existing functionality just to simplify the implementation.
14. Do not weaken security controls to make tests pass.
15. Never claim an issue is fixed without verification.

## 4. Authentication-Specific Requirements

Give special attention to authentication because it is a critical part of this project.

Verify that:

- Passwords are securely hashed.
- Password verification is implemented correctly.
- JWT signing algorithms are explicitly validated.
- JWT expiration is enforced.
- Invalid, expired, or malformed tokens are rejected.
- Authentication middleware correctly identifies users.
- Protected endpoints cannot be accessed without valid authentication.
- Authorization checks are performed where required.
- User information is not unnecessarily exposed.
- Refresh token handling is secure, if implemented.
- Logout and token revocation behave correctly, if implemented.
- Authentication failures return appropriate HTTP status codes.

Add tests for valid tokens, invalid tokens, expired tokens, missing tokens, malformed tokens, and unauthorized access.

Do not change the authentication flow unnecessarily if the existing implementation is already secure.

## 5. Testing & Validation

After each implementation phase:

1. Run gofmt on modified Go files.
2. Run relevant unit tests.
3. Run `go vet ./...`.
4. Run `go test ./...`.
5. Run `go test -race ./...` where applicable.
6. Check compilation and dependency consistency.
7. Verify that existing API behavior has not been unintentionally broken.

If tests fail:
- Investigate the actual root cause.
- Fix the implementation if the failure is caused by your changes.
- Do not delete or weaken tests just to obtain a passing result.
- Clearly report pre-existing failures separately.

If database or external infrastructure is unavailable, do not pretend integration tests have passed.

## 6. Documentation Updates

Update `docs/GO_BACKEND_AUDIT.md` as implementation progresses.

For every finding, record:

- Finding ID.
- Original severity.
- Implementation status: FIXED / PARTIALLY FIXED / NOT FIXED / INVALID.
- Files modified.
- Description of the changes.
- Tests performed.
- Verification results.
- Remaining risks.

Keep the original audit findings for traceability.

## 7. Execution Strategy

Work autonomously and systematically.

Do not stop after generating an implementation plan.

Do not ask for confirmation between phases unless:
- A destructive database operation is required.
- A breaking API change is unavoidable.
- Required credentials or external infrastructure are unavailable.
- A decision requires business context that cannot be inferred safely.

If a finding is invalid or cannot be fixed safely, skip it and explain why.

Do not silently ignore any finding.

Complete as many safe and verifiable improvements as possible.

## 8. Final Deliverables

After implementation, provide:

### A. Implementation Summary
Explain what has actually been changed.

### B. Findings Status
Provide a table:

| Finding ID | Severity | Status | Description |
|---|---|---|---|

### C. Files Modified
List important modified files and explain their purpose.

### D. Testing Results
Report actual execution results:

| Command | Status | Notes |
|---|---|---|
| gofmt | PASS/FAIL | |
| go vet ./... | PASS/FAIL | |
| go test ./... | PASS/FAIL | |
| go test -race ./... | PASS/FAIL/SKIPPED | |

### E. Remaining Issues
List unresolved findings and explain why they remain unresolved.

### F. Production Readiness Assessment
Provide an honest assessment of the project's current state.

## 9. README & API Documentation

As part of the implementation, create or update the project's README and Swagger/OpenAPI documentation.

### A. README.md

Create a comprehensive `README.md` at the project root.

The documentation must reflect the actual project implementation.

Include:

1. **Project Overview**
   - Project description.
   - Main features.
   - Technology stack.
   - Architecture overview.

2. **Prerequisites**
   - Required Go version based on go.mod.
   - Database requirements.
   - Redis or other infrastructure requirements, if applicable.
   - Required development tools.

3. **Installation & Setup**
   - Clone repository.
   - Install dependencies.
   - Configure environment variables.
   - Database setup.
   - Database migration.
   - Running the application locally.

4. **Environment Variables**

   Provide a complete environment variable reference:

   | Variable | Description | Required | Default |
   |---|---|---|---|

   Inspect the actual configuration implementation and list all supported variables.

   Never expose real credentials, tokens, or production secrets. Use safe example values.

5. **Project Structure**

   Explain the important directories and their responsibilities.

6. **Authentication**
   - Authentication mechanism.
   - Login flow.
   - Token handling.
   - Protected endpoints.
   - Authorization mechanism, if implemented.

7. **API Overview**
   - API base URL.
   - Available endpoint groups.
   - Authentication requirements.
   - Response format.
   - Error response format.
   - Pagination conventions, if applicable.

8. **Running Tests**
   - Unit tests.
   - Integration tests.
   - Race detection.
   - Static analysis.

9. **Development Guidelines**
   - Code formatting.
   - Naming conventions.
   - Adding new endpoints.
   - Adding middleware.
   - Database migration workflow.

10. **Deployment**
    - Production build.
    - Environment configuration.
    - Database migration.
    - Running the application.
    - Health check endpoints, if available.

11. **License**
    - Document the actual license if present.
    - Do not invent a license.

### B. Swagger / OpenAPI Documentation

Generate API documentation based on the actual HTTP routes, handlers, request DTOs, response DTOs, middleware, and validation rules.

Use Swagger UI with OpenAPI 3.0 or a compatible specification.

Preferred implementation:
- OpenAPI 3.0 specification.
- Swagger UI for interactive API exploration.
- Bearer JWT authentication scheme if JWT is used.
- Reusable schemas for requests and responses.

Before choosing a Swagger generation tool, inspect the current project dependencies and Go version. Prefer a tool that integrates cleanly with the existing Chi architecture.

Do not introduce unnecessary dependencies.

Document all implemented endpoints, including:

- HTTP method.
- Endpoint path.
- Summary.
- Description.
- Tags.
- Path parameters.
- Query parameters.
- Request headers.
- Request body.
- Response schemas.
- Success status codes.
- Error status codes.
- Authentication requirements.

Pay special attention to authentication endpoints:

- Register, if implemented.
- Login.
- Refresh token, if implemented.
- Logout.
- Current authenticated user / profile.
- Password reset, if implemented.

For protected endpoints, document the required Bearer token authentication.

### C. Swagger UI Integration

If technically appropriate, expose Swagger UI through a dedicated route, for example:

`/swagger/index.html`

Requirements:

- Swagger UI must load the generated OpenAPI specification.
- API documentation must be accessible in the development environment.
- Production exposure should be configurable.
- Do not expose internal secrets or sensitive implementation details.
- Ensure the documentation route does not accidentally bypass application security policies.

### D. API Documentation Accuracy

Follow these rules:

1. Inspect every registered Chi route.
2. Verify the actual handler behavior.
3. Use real request and response structures.
4. Do not invent endpoints.
5. Do not invent request fields.
6. Do not document unsupported status codes.
7. Ensure documented authentication requirements match actual middleware.
8. Keep documentation synchronized with the implementation.
9. Include realistic but fictional request and response examples.
10. Clearly document optional versus required fields.

### E. Documentation Validation

After generating documentation:

- Verify that the OpenAPI specification is valid.
- Verify that Swagger UI loads successfully.
- Verify that documented routes match actual registered routes.
- Verify request and response schemas.
- Verify JWT security definitions.
- Check that all relevant endpoints are documented.

### F. Expected Deliverables

The final project should include:

```text
project-root/
├── README.md
├── docs/
│   ├── GO_BACKEND_AUDIT.md
│   └── openapi.yaml
└── ...
```

The OpenAPI file location may be adjusted if the chosen Swagger tooling requires a different structure.

Update README.md with:

- Swagger UI URL.
- OpenAPI specification location.
- Instructions for generating or updating documentation.

Update `docs/GO_BACKEND_AUDIT.md` with the documentation implementation status and validation results.

**Important:** Documentation must describe the final implemented code, not the original implementation or an imagined architecture.

## FINAL OBJECTIVE

Transform the existing Go Chi backend into a more secure, reliable, maintainable, and production-ready application.

Prioritize correctness and security over unnecessary architectural complexity.

**Do not just tell me what should be done. Actually implement the improvements, test them, and document the results.**