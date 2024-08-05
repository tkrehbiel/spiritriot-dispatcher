This microservice is designed to subscribe to an AWS SNS publication that is triggered when a new post is added to a blog.

It fetches a post, scans it for external links, and publishes a message to another topic for each link that's found. Subscribers of the topic can send webmentions and/or pingback notifications for each link.

Incoming Message Body:
- post url
- published date (iso 8601 date)
- detected date (iso 8601 date)

Outgoing Message Body:
- source url
- target url
