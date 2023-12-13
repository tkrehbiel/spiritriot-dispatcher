This microservice is designed to subscribe to an AWS SNS publication that is triggered when a new post is added to a blog.

It fetches the post, scans it for external links, and publishes a message to another topic for each link that's found. Subscribers of the topic can send webmentions and/or pingback notifications for each link.

Message Body:
- url
- published (iso 8601 date)
- detected (iso 8601 date)
