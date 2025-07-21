#!/usr/bin/env python3

# This python file runs a http server that acts as the ingest for all log
# and spec files sent to it and queues the files for AI processing.
#
# Author: Chris Mayenschein
# GitHub: https://github.com/cmayen/paila
# Date: 2025-07-21
# Last Modified: 2025-07-21
#
# Usage: python3 paila-ingest-py.py
#
# #############################################################################


import os
import re
import json
import socket
from http.server import HTTPServer, BaseHTTPRequestHandler
from urllib.parse import urlparse
from email.parser import BytesParser
from email.policy import default


# override listenPort with PAILA_INGEST_PORT env var in main func
listen_port = int(os.getenv("PAILA_INGEST_PORT", 8181))

uploads_dir = os.path.expanduser("/.paila-ingest/uploads")
reports_dir = os.path.expanduser("/.paila-ingest/reports")
archive_dir = os.path.expanduser("/.paila-ingest/archive")

# Ensure directories exist
os.makedirs(uploads_dir, exist_ok=True)
os.makedirs(reports_dir, exist_ok=True)
os.makedirs(archive_dir, exist_ok=True)

# get the outbound up so we can output debug info on start
def get_local_outbound_ip():
    try:
        s = socket.socket(socket.AF_INET, socket.SOCK_DGRAM)
        s.connect(("8.8.8.8", 80))
        ip = s.getsockname()[0]
        s.close()
        return ip
    except Exception:
        return "127.0.0.1"

# handler for the file upload
class SimpleUploadHandler(BaseHTTPRequestHandler):

    def send_json(self, data, code=200):
        self.send_response(code)
        self.send_header("Content-Type", "application/json")
        self.end_headers()
        self.wfile.write(json.dumps(data).encode("utf-8"))

    def do_POST(self):
        parsed_path = urlparse(self.path)
        if parsed_path.path != "/uploadlog":
            self.send_json({
                "status": "404",
                "message": "Not Found"
            }, 404)
            return
        content_type = self.headers.get("Content-Type")
        if not content_type or not content_type.startswith("multipart/form-data"):
            self.send_json({
                "status": "400",
                "message": "Expected multipart/form-data"
            }, 400)
            return

        content_length = int(self.headers.get("Content-Length", 0))
        body = self.rfile.read(content_length)

        # Parse multipart form data using email.parser
        msg = BytesParser(policy=default).parsebytes(
            b"Content-Type: " + content_type.encode() + b"\r\n\r\n" + body
        )

        fields = {}
        for part in msg.iter_parts():
            name = part.get_param('name', header='content-disposition')
            if name:
                if part.get_content_type() == "text/plain":
                    fields[name] = [part.get_content().strip()]
                else:
                    fields[name] = [part.get_payload(decode=True)]

        # sanitizing
        # remove all non-alphanumeric characters except hyphens and periods
        host = self.clean_field(fields.get("host", [""])[0].decode() if isinstance(fields.get("host", [""])[0], bytes) else fields.get("host", [""])[0])
        date = self.clean_field(fields.get("date", [""])[0].decode() if isinstance(fields.get("date", [""])[0], bytes) else fields.get("date", [""])[0])

        # make sure log is sent
        if "log" not in fields or len(fields["log"]) == 0:
            self.send_json({
                "status": "400",
                "message": "Missing file field 'log'",
                "host": host,
                "date": date
            }, 400)
            return

        # The filename comes from Content-Disposition, which isn't 
        # retained here, so we use a default
        filename = f"{host}_{date}.log"
        filepath = os.path.join(uploads_dir, filename)

        # Create a new file on the filesystem to save the uploaded content
        try:
            log_data = fields["log"][0]
            if isinstance(log_data, str):
                log_data = log_data.encode("utf-8")
            with open(filepath, "wb") as f:
                f.write(log_data)
        except Exception as e:
            self.send_json({
                "status": "500",
                "message": f"Error writing file: {str(e)}",
                "filename": filename,
                "host": host,
                "date": date,
                "filepath": filepath
            }, 500)
            return

        # send the success response
        self.send_json({
            "status": "201",
            "message": "File uploaded successfully",
            "filename": filename,
            "host": host,
            "date": date,
            "filepath": filepath
        }, 201)
        

    # make sure it is a post method
    def do_GET(self):
        self.send_json({
            "status": "405",
            "message": "Only POST to /uploadlog is supported"
        }, 405)

    def clean_field(self, value):
        return re.sub(r"[^a-zA-Z0-9.-]", "", value)

def run():
    ip = get_local_outbound_ip()
    print(f"Server listening on :{listen_port}")
    print(f"  ./paila-logpush.sh -u http://{ip}:{listen_port}/uploadlog")
    httpd = HTTPServer(("", listen_port), SimpleUploadHandler)
    httpd.serve_forever()

if __name__ == "__main__":
    run()
