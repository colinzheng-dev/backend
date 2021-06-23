# Email service

## Design

The idea here is to provide an email service that can send emails
based on templates that Tiago and Agi can create, with email sending
driven off of events that happen in the backend.

This means that the email service needs three things do its job:

1. A means of sending email. I looked at a few options here, but the
   simplest approach (that offers templating web app that
   non-technical people can use) seems to be
   [Mailjet](https://mailjet.com).

2. A set of templates. Mailjet stores these, and we'll trigger sending
   them based on a JSON message delivered from a Pub/Sub subscription.

2. A set of event-to-email mappings saying which event types on which
   Pub/Sub topics cause the sending of emails from which templates.
   To start with, these can be simple topic, event type, template
   triples, although we might want to add other features in the future
   like the possibility to aggregate multiple events into single
   "digest" emails.

NOTE: Google Pub/Sub can deliver messages multiple times, so the email
service needs to keep a record of message IDs that have been
processed, at least for a time period corresponding to the message
delivery timeout.

## Configuration

 - PROJECT_ID: Google project ID. (Setting this to "dev" lets you run
   locally for development, which just causes emails to be printed to
   the console instead of sent.)
 - DATABASE_URL: database for storing processed message IDs.
 - PORT: server port.
 - CREDENTIALS_PATH: path to GCP service account credentials.
 - MAILJET_API_KEY_PUBLIC, MAILJET_API_KEY_PRIVATE: API keys for
   Mailjet.
 - SIMULTANEOUS_EMAILS: maximum number of emails to send
   simultaneously.

## Email sending process

 - Email service subscribes to events.
 - Pub/Sub event arrives.
