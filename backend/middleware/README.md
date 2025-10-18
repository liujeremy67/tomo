# Middleware Overview

In Go, every HTTP request is handled by a handler, a function that receives two things:

1. **ResponseWriter (`w`)** – Provided by Go; lets middleware or handlers send responses back. The “outbox” for sending data back to the client.
2. **Request (`r`)** – Provided by Go; the “inbox” containing all information sent by the client, including the JSON body.

Middleware is a wrapper around a handler. It can:

- Inspect or modify the request before the handler runs.
- Inspect or modify the response after the handler runs.
- Perform additional tasks (logging, authentication, rate limiting, etc.) without changing the handler itself.

---

## Logger

1. **Go receives a request**  
   - The server creates a `Request` object representing all the incoming data.  
   - It also provides a `ResponseWriter` object to collect the outgoing response.

2. **Logger intercepts the request**  
   - Logger wraps the original handler.  
   - When the request reaches Logger, it has access to both the `Request` (inbox) and `ResponseWriter` (outbox).

3. **Pre-processing**  
   - Logger records the current time.

4. **Delegate to the next handler**  
   - Logger calls the original handler, passing along the same inbox and outbox.  
   - The original handler processes the request and writes the response.

5. **Post-processing**  
   - After the handler finishes, Logger calculates the elapsed time.  
   - It logs the HTTP method, path, and duration.  

---

## Rate Limiter
Similar to Logger; function that wraps a handler to perform pre- and post-processing around requests. Prevents making too many requests too quickly.

1. **Go receives a request**  
   - For each incoming request, Go creates a `Request` object representing the client’s data.  
   - It also provides a `ResponseWriter` object to send the response.

2. **RateLimit intercepts the request**  
   - RateLimit wraps the original handler.  
   - It receives the request (`r`) and response (`w`) objects.

3. **Check the client’s last request time**  
   - The middleware identifies the client by IP (`r.RemoteAddr`).  
   - It looks up when this client last made a request.  
   - If the last request was too recent (e.g., less than 1 second ago), it stops the request and returns a `429 Too Many Requests` response.

4. **Allow or reject the request**  
   - If the request is allowed, it updates the client’s last request timestamp and calls the original handler.  
   - If rejected, it responds immediately without calling the handler.

---