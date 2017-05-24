HTTP interface for email sending
================================

## Configuration

    [graylog]
    address = logs.acme.com:12212

    [http]
    address = :8899

    [smtp]
    host = smtp.acme.com
    port = 465
    user = bigboss@acme.com
    password = tH3v3rYs3cr37p4sSw0rD

## Mail sending

    curl -XPOST 'http://rest-mail.acme.com:8899/' \
     -H "X-To: bigboss@acme.com" \
     -H "X-From: bob@acme.com" \
     -H "X-Subject: Quarter report" \
     -H "Content-Type: text/html" \
     -d "Check out this link buddy <a href='http://acme.com'>Hot chicks</a>"

## Warning!

HTTP server has no authentication mechanism so NEVER EVER open it in public.