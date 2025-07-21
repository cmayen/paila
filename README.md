# paila
 **P**arsing **AI** for **L**ogging & **A**nalysis ("PY‑luh")

A smart agent tool that watches systems, understands logs, finds problems, and suggests solutions.

🧠  Logs In. Insights Out.  AI-Powered Clarity for Your Servers.

---

2025-07-21
This project is in its infancy, so there will be many changes as areas are completed.

---


### Docker Containers


#### paila-ollama-rocm

An ollama container built with rocm in order to work with AMD chipsets.


#### paila-ingest

The ingest server that will receive log files from remote machines and queue the files for AI analysis.


### Log Discovery Tool

#### paila-logpush.sh

Shell script copied to remote servers that filters log files and interacts with journalctl (if present) to generate an ingest file for the paila-ingest server.


### Ubuntu Installer Chain

docker/install-docker-amd-ubuntu.sh
docker/paila-ollama-rocm/paila-ollama-rocm-install.sh
docker/paila-ingest/image/paila-ingest-image-create.sh # development
docker/paila-ingest/paila-ingest-install.sh



