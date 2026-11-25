# External APIs

The template does not include a preconfigured external API client or credential
store. For privileged API requests, prefer Go's net/http with a timeout,
context cancellation, bounded responses, and typed request/response models.
Expose narrowly scoped service methods and regenerate TypeScript bindings.
Do not create a generic arbitrary URL or shell execution bridge.

Frontend fetch is appropriate for public endpoints that permit browser
requests (including CORS). Keep private credentials out of frontend assets,
localStorage and logs. Add a suitable OS credential store when the application
needs persisted secrets; ordinary JSON preferences are not secret storage.

Handle timeouts, cancellation, non-success status codes and malformed data.
Avoid automatically retrying non-idempotent writes. Add tests using httptest
and mock the generated bindings in frontend tests.
