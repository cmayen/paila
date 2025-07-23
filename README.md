# paila
 **P**arsing **AI** for **L**ogging & **A**nalysis ("PY‑luh")

A smart tool that watches systems, understands logs, finds problems, and suggests solutions.

🧠  Logs In. Insights Out.  AI-Powered Clarity for Your Servers.

---

2025-07-21
This part time project is in its infancy, so there will be many changes as areas are completed. Things may get ugly. At the moment this project is in prototpe phase and being developed in a secure environment. Further security precautions and checks will be integrated as development progresses.

---


## Docker Containers


#### paila-ollama

The AI / LLM ollama container built with the ollama/ollama:rocm image in order to work with AMD chipsets, and should use the ollama/ollama:latest for NVIDIA chipsets.


#### paila-ingest

The ingest server that will receive log files from remote machines running the paila-logpush.sh shell script and queue the files for AI analysis.


#### paila-reporter

The reporter server calls for AI analysis and provides a web interface. View logs, reports, and ability to manually re-generate reports.


---


## Log Discovery Tool

#### paila-logpush.sh

Shell script copied to remote servers that filters log files for unique entries and interacts with journalctl (if present) to generate and upload an ingest file for the paila-ingest server containing information about the previous day.


---


![image of web ui report](https://github.com/cmayen/paila/blob/main/.readme-assets/paila-250722.png?raw=true)
