# Logger Middleware — Conceptual Overview

In Go, every HTTP request is handled by a handler, a function that receives two things:

1. **ResponseWriter (`w`)** – the “outbox” for sending data back to the client.
2. **Request (`r`)** – the “inbox” containing all information sent by the client.

Middleware, like Logger, is a wrapper around a handler. It can:

- Inspect or modify the request before the handler runs.
- Inspect or modify the response after the handler runs.
- Perform additional tasks (logging, authentication, rate limiting, etc.) without changing the handler itself.

---

## How Logger Works Conceptually

1. **Go receives a request**  
   - The server creates a `Request` object representing all the incoming data.  
   - It also provides a `ResponseWriter` object to collect the outgoing response.

2. **Logger intercepts the request**  
   - Logger wraps the original handler.  
   - When the request reaches Logger, it has access to both the `Request` (inbox) and `ResponseWriter` (outbox).

3. **Pre-processing**  
   - Logger records the current time.  
   - Conceptually, it says: “I’ll measure how long this request takes.”

4. **Delegate to the next handler**  
   - Logger calls the original handler, passing along the same inbox and outbox.  
   - The original handler processes the request and writes the response.

5. **Post-processing**  
   - After the handler finishes, Logger calculates the elapsed time.  
   - It logs the HTTP method, path, and duration.  

---
