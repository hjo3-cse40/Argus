# Data Collection Research - Collector Team

**Team:** Sam, Bill  
**Date:** February 2026  
**Sprint:** Sprint 1

## Overview

This document outlines research on various methods for collecting data/events that will feed into the Argus system. The goal is to identify and evaluate different approaches for ingesting events from various sources.

## Event Sources to Consider

### 1. Web Scraping / Crawling
**Description:** Automated collection of data from websites  
**Use Cases:**
- News articles
- Product listings
- Job postings
- Social media posts (public)

**Technologies:**
- **Go libraries:**
  - `goquery` - HTML parsing
  - `colly` - Web scraping framework
  - `chromedp` - Headless Chrome automation
- **Python alternatives:**
  - Scrapy
  - BeautifulSoup
  - Selenium

**Pros:**
- Direct access to public data
- No API dependencies
- Can target any website

**Cons:**
- Legal/ethical considerations (robots.txt, ToS)
- Fragile (breaks when sites change)
- Rate limiting concerns
- Requires maintenance

**Implementation Notes:**
- Respect robots.txt
- Implement rate limiting
- Handle dynamic content (JS rendering)
- Error handling for site changes

---

### 2. RSS/Atom Feeds
**Description:** Subscribe to RSS/Atom feeds for content updates  
**Use Cases:**
- News sites
- Blogs
- Podcasts
- Newsletters

**Technologies:**
- **Go libraries:**
  - `github.com/mmcdole/gofeed` - RSS/Atom parser
  - `github.com/SlyMarbo/rss` - RSS parser

**Pros:**
- Standardized format
- Easy to implement
- Many sites support it
- Efficient (only new items)

**Cons:**
- Not all sites have feeds
- Some feeds incomplete
- Polling required (not real-time)

**Implementation Notes:**
- Polling interval configuration
- Deduplication (check if already processed)
- Handle feed format variations

---

### 3. API Integrations
**Description:** Integrate with third-party APIs  
**Use Cases:**
- Social media APIs (Twitter/X, Reddit, etc.)
- News APIs (NewsAPI, Guardian API)
- GitHub events
- HackerNews API

**Technologies:**
- Standard HTTP clients
- OAuth for authenticated APIs
- Webhooks (if supported)

**Pros:**
- Official, stable interfaces
- Often real-time (webhooks)
- Structured data
- Rate limits documented

**Cons:**
- API keys required
- Rate limits
- Costs (some APIs are paid)
- API changes can break integration

**Implementation Notes:**
- API key management
- Rate limit handling
- Retry logic
- Webhook vs polling decision

---

### 4. Webhooks
**Description:** Receive events via HTTP callbacks  
**Use Cases:**
- GitHub webhooks
- Stripe events
- Custom integrations
- Internal services

**Technologies:**
- HTTP server endpoint
- Signature verification
- Event queuing

**Pros:**
- Real-time
- Push-based (efficient)
- No polling needed
- Standard HTTP

**Cons:**
- Requires public endpoint (or tunnel)
- Security considerations
- Need to handle downtime

**Implementation Notes:**
- Webhook endpoint in API
- Signature verification
- Idempotency handling
- Queue events immediately

---

### 5. File Monitoring
**Description:** Watch directories for new files  
**Use Cases:**
- Log file processing
- Batch imports
- Data exports from other systems
- CSV/JSON file drops

**Technologies:**
- **Go libraries:**
  - `github.com/fsnotify/fsnotify` - File system notifications
  - `github.com/radovskyb/watcher` - File watcher

**Pros:**
- Simple for batch processing
- No external dependencies
- Good for legacy systems

**Cons:**
- Not real-time (polling-based)
- File format parsing required
- Error handling for malformed files

**Implementation Notes:**
- Watch specific directories
- Parse file formats (CSV, JSON, etc.)
- Handle file locking
- Archive processed files

---

### 6. Database Change Data Capture (CDC)
**Description:** Monitor database changes  
**Use Cases:**
- Replicate data from other systems
- Monitor database events
- Sync between databases

**Technologies:**
- Postgres logical replication
- MySQL binlog
- Debezium (Kafka Connect)

**Pros:**
- Real-time database changes
- No application changes needed
- Reliable

**Cons:**
- Complex setup
- Database-specific
- Performance overhead

**Implementation Notes:**
- Postgres logical replication
- Parse WAL events
- Transform to event format

---

### 7. Message Queue Subscriptions
**Description:** Subscribe to existing message queues  
**Use Cases:**
- Integrate with other microservices
- Event-driven architecture
- Legacy system integration

**Technologies:**
- RabbitMQ
- Kafka
- NATS
- Redis Streams

**Pros:**
- Decoupled architecture
- Scalable
- Reliable delivery

**Cons:**
- Requires access to queue
- Queue-specific implementation

**Implementation Notes:**
- RabbitMQ consumer (already have infrastructure)
- Kafka consumer (if needed)
- Message format standardization

---

### 8. Email Parsing
**Description:** Parse emails for events  
**Use Cases:**
- Newsletter subscriptions
- Email notifications
- Automated reports

**Technologies:**
- IMAP/POP3 clients
- Email parsing libraries
- **Go libraries:**
  - `github.com/emersion/go-imap`
  - `github.com/jhillyerd/enmime` - Email parsing

**Pros:**
- Universal (everyone has email)
- Can subscribe to email lists

**Cons:**
- Complex parsing
- Spam handling
- Rate limits

**Implementation Notes:**
- IMAP connection
- Parse email content
- Extract links/articles
- Handle attachments

---

## Recommended Approach for Sprint 1

For the initial implementation, we recommend starting with:

1. **RSS/Atom Feeds** - Easy to implement, many sources available
2. **Webhooks** - Add webhook endpoint to API for external integrations
3. **CLI Tool** - Already implemented for testing/synthetic events

## Future Considerations

- **Rate Limiting:** Implement per-source rate limiting
- **Deduplication:** Track processed items to avoid duplicates
- **Error Handling:** Retry logic, dead letter queues
- **Monitoring:** Track collection success/failure rates
- **Configuration:** Per-source configuration (URLs, credentials, etc.)

## Next Steps

1. Implement RSS feed collector
2. Add webhook endpoint to API
3. Create collector service architecture
4. Design source configuration system
5. Implement rate limiting and deduplication

## References

- [Go Colly Documentation](https://go-colly.org/)
- [GoFeed RSS Parser](https://github.com/mmcdole/gofeed)
- [fsnotify File Watcher](https://github.com/fsnotify/fsnotify)
- [RabbitMQ Best Practices](https://www.rabbitmq.com/best-practices.html)
