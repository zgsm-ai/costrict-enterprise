#!/usr/bin/env python3
"""Send a signed GitHub Stars event using only the Python standard library."""

import argparse
import getpass
import hashlib
import hmac
import json
import os
import sys
import urllib.error
import urllib.parse
import urllib.request
import uuid


class NoRedirect(urllib.request.HTTPRedirectHandler):
    def redirect_request(self, req, fp, code, msg, headers, newurl):
        return None


def main():
    parser = argparse.ArgumentParser(description=__doc__)
    parser.add_argument("--url", default=os.getenv("WEBHOOK_URL"),
                        help="Full webhook URL (or WEBHOOK_URL)")
    parser.add_argument("--owner", default=os.getenv("GITHUB_OWNER", "zgsm-ai"))
    parser.add_argument("--repo", default=os.getenv("GITHUB_REPO", "costrict"))
    parser.add_argument("--user-id", type=int, default=os.getenv("GITHUB_USER_ID"),
                        help="Positive numeric GitHub user ID (or GITHUB_USER_ID)")
    parser.add_argument("--login", default="webhook-test", help="Login used in server logs")
    parser.add_argument("--action", choices=("created", "deleted"),
                        default=os.getenv("STAR_ACTION", "created"))
    args = parser.parse_args()
    if not args.url:
        parser.error("--url or WEBHOOK_URL is required")
    url = urllib.parse.urlsplit(args.url)
    if url.scheme not in ("http", "https") or not url.netloc:
        parser.error("URL must be an absolute http:// or https:// URL")
    if args.user_id is None or args.user_id <= 0:
        parser.error("--user-id or GITHUB_USER_ID must be a positive integer")
    if not args.owner or not args.repo:
        parser.error("owner and repo must not be empty")
    if args.action not in ("created", "deleted"):
        parser.error("STAR_ACTION must be created or deleted")

    secret = os.getenv("WEBHOOK_SECRET")
    if not secret:
        if not sys.stdin.isatty():
            parser.error("set WEBHOOK_SECRET when running non-interactively")
        secret = getpass.getpass("Webhook Secret: ")
    if not secret:
        parser.error("Webhook Secret must not be empty")

    payload = {
        "action": args.action,
        "repository": {"name": args.repo, "owner": {"login": args.owner}},
        "sender": {"id": args.user_id, "login": args.login},
    }
    body = json.dumps(payload, separators=(",", ":")).encode()
    signature = hmac.new(secret.encode(), body, hashlib.sha256).hexdigest()
    delivery = str(uuid.uuid4())
    request = urllib.request.Request(args.url, data=body, method="POST", headers={
        "Content-Type": "application/json",
        "X-GitHub-Event": "star",
        "X-GitHub-Delivery": delivery,
        "X-Hub-Signature-256": "sha256=" + signature,
    })
    print(f"Sending {args.action}: {args.owner}/{args.repo}, user ID {args.user_id}")
    print(f"Delivery: {delivery}")
    try:
        opener = urllib.request.build_opener(NoRedirect())
        with opener.open(request, timeout=15) as response:
            print(f"HTTP {response.status}")
            print(response.read().decode(errors="replace"))
            return 0 if response.status == 204 else 1
    except urllib.error.HTTPError as error:
        with error:
            print(f"HTTP {error.code}", file=sys.stderr)
            print(error.read().decode(errors="replace"), file=sys.stderr)
        return 1
    except (urllib.error.URLError, OSError) as error:
        print(f"Request failed: {error}", file=sys.stderr)
        return 1


if __name__ == "__main__":
    sys.exit(main())
