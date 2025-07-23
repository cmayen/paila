#!/bin/bash
#
# This script scans the log folders and interacts with journalctl
# to collect issues about the system, then uploads the log data
# to a handling server for further processing. If logs are found,
# this script will also include a report gathering system information
# to aid with troubleshooting.
#
# Author: Chris Mayenschein
# GitHub: https://github.com/cmayen/paila
# Date: 2025-07-20
# Last Modified: 2025-07-23
#
# Usage: ./paila-logpush.sh
# Usage: ./paila-logpush.sh [-u output_url] [-d out_directory] [-l log_directory]
# Example: paila-logpush.sh -u http://localhost:8181/uploadlog -l /var/log
#
################################################################################


# curl http file push location
# check for PAILA_OUTURL environment variable
if [[ -z "${PAILA_OUTURL}" ]]; then
  # not defined, set default
  OUTPUTURL="http://localhost:8181/uploadlog"
else
  OUTPUTURL="${PAILA_OUTURL}"
fi


# log file generation location
# check for PAILA_OUTDIR environment variable
if [[ -z "${PAILA_OUTDIR}" ]]; then
  # not defined, set default
  OUTPUTDIR="/var/tmp"
else
  OUTPUTDIR="${PAILA_OUTDIR}"
fi


# system log directory location
# check for PAILA_LOGDIR environment variable
if [[ -z "${PAILA_LOGDIR}" ]]; then
  # not defined, set default
  LOGDIR="/var/log"
else
  LOGDIR="${PAILA_LOGDIR}"
fi


# get passed in options
# passed options will override env variables
#   'u:' output url
#   'd:' output directory
#   'l:' log directory
while getopts "u:d:l:" opt; do
  case $opt in

    u) # output url curl with call to
      OUTPUTURL=$OPTARG;;

    d) # output directory for generated file
      OUTPUTDIR=$OPTARG;;

    l) # log directory containing the logs to scan
      LOGDIR=$OPTARG;;

    \?) # Handle invalid options
      echo "Usage: $0 [-u output_url] [-d out_directory] [-l log_directory]" >&2
      exit 1;;
  esac
done


# make sure the output and log directories end with a trailing /
if [[ "$OUTPUTDIR" != */ ]]; then
  # If not, append /
  OUTPUTDIR="${OUTPUTDIR}/"
fi
if [[ "$LOGDIR" != */ ]]; then
  # If not, append /
  LOGDIR="${LOGDIR}/"
fi


# allows the last command in a pipeline to run in the current shell,
# thus allowing variable changes within a while loop to persist
shopt -s lastpipe # Enable lastpipe option


# setup some date objects for filtering and naming
DATE_S=$(date --date="yesterday"  +"%Y-%m-%d")
DATE_Y=$(date --date="yesterday"  +"%Y-%m-%d 00:00:00")
DATE_T=$(date --date="today"  +"%Y-%m-%d 00:00:00")


# regex pattern match for date string existance
DATE_P="[0-9]{4}-[0-9]{2}-[0-9]{2}"


# scan the log dir
LOG_FILES=$(find $LOGDIR -name "*.log" -mtime -3)


# determine the hostname
HOST=$(cat /etc/hostname)


# set output path for generated logs file
OUTPUTPATH="${OUTPUTDIR}${HOST}--${DATE_S}.logs.txt"


# init logsfound to 0, change to 1 if anything comes up
# so the script will either gather further system information
# or generate a "good health" report in the logs file
LOGSFOUND=0


# a generic header for the logged issues
echo -e "\n============================================" > "${OUTPUTPATH}"
echo -e "= Begin Logged Issues Report" >> "${OUTPUTPATH}"
echo -e "= Host: ${HOST}" >> "${OUTPUTPATH}"
echo -e "= Date: ${DATE_S}" >> "${OUTPUTPATH}"
echo -e "============================================" >> "${OUTPUTPATH}"


# loop the found log files
for LOGFILE in $LOG_FILES; do


  # get 1 possible error back from the file, this content will
  # be used to determine whether to dig deeper, and to check if
  # a date format is available to filter by
  grep -Eiw -m 1 "warning|error|critical|alert|fatal" $LOGFILE | while read -r checkline; do


    # check if a date filter can be used
    if [[ "$checkline" =~ $DATE_P ]]; then


      # date is available for filtering, get a count of the errors with the new filter applied
      LOGCOUNT=$(grep -Eiw "warning|error|critical|alert|fatal" $LOGFILE | grep -c "$DATE_S")


      # if the count is greater than 0 then there are logs to report
      if [ "$LOGCOUNT" -gt 0 ]; then
        LOGSFOUND=1

        # date filtered log entries exist
        echo -e "\n======================" >> "${OUTPUTPATH}"
        echo -e "======================" >> "${OUTPUTPATH}"
        echo -e "=== Log: ${LOGFILE}" >> "${OUTPUTPATH}"
        echo -e "=======" >> "${OUTPUTPATH}"
        # do the output search
        grep -Eiw "warning|error|critical|alert|fatal" $LOGFILE | grep "$DATE_S" | sort -u | while read -r line; do
          #
          echo "${line}" >> "${OUTPUTPATH}"
          #
        done
      fi
    else
      LOGSFOUND=1
      # no date found to filter by
      echo -e "\n======================" >> "${OUTPUTPATH}"
      echo -e "======================" >> "${OUTPUTPATH}"
      echo -e "=== Log: ${LOGFILE}" >> "${OUTPUTPATH}"
      echo -e "=======" >> "${OUTPUTPATH}"
      grep -Eiw "warning|error|critical|alert|fatal" $LOGFILE | sort -u | while read -r line; do
        #
        echo "$line" >> "${OUTPUTPATH}"
        #
      done
    fi
  done

done

# Call to journalctl requesting critical information from yesterday.
# store the data in a value so we can suppress the log header if there
# is nothing returned
JCTL=$(journalctl -S "$DATE_Y" -U "$DATE_T" --no-pager --priority=3..0)
# check for no entries response
if [[ ! "$JCTL" == "-- No entries --" ]] && [[ ! "$JCTL" == "" ]]; then
  # we have entries, send the header and LOGSFOUND
  LOGSFOUND=1
  echo -e "\n======================" >> "${OUTPUTPATH}"
  echo -e "======================" >> "${OUTPUTPATH}"
  echo -e "=== Log: journalctl -S \"${DATE_Y}\" -U \"${DATE_T}\" --no-pager --priority=3..0  ===" >> "${OUTPUTPATH}"
  echo -e "=======" >> "${OUTPUTPATH}"
  echo "$JCTL" >> "${OUTPUTPATH}"
fi

# check if there were no logs found
if [ "$LOGSFOUND" -eq 0 ]; then
  echo -e "\n-- No Entries -- System is known to be in good health. --" >> "${OUTPUTPATH}"
fi

# a generic footer for the logged issues
echo -e "\n============================================" >> "${OUTPUTPATH}"
echo -e "= End Logged Issues Report" >> "${OUTPUTPATH}"
echo -e "============================================\n" >> "${OUTPUTPATH}"

# The system information report is causing issues with the ollama so it is disabled for now
# set the -eq check to 1 to always run the system information report if logs are found

# check if there were any logs found
if [ "$LOGSFOUND" -eq 10 ]; then
  # logs found, retreiving system information

  COMMANDS=(
    "hostnamectl" 
    "hostname -I"
    "ip a"
    "lsblk --all --output-all"
    "lsusb --verbose"
    "lspci -v"
    "free -m"
    "cat /proc/meminfo"
    "cat /proc/cpuinfo"
    "lscpu --all --extended --output-all"
  )

  echo -e "\n============================================" >> "${OUTPUTPATH}"
  echo -e "= Begin System Information Report" >> "${OUTPUTPATH}"
  echo -e "= Host: ${HOST}" >> "${OUTPUTPATH}"
  echo -e "= Date: ${DATE_S}" >> "${OUTPUTPATH}"
  echo -e "============================================" >> "${OUTPUTPATH}"

  # double quotes around "${array[@]}" are really important. Without 
  # them, the for loop will break up the array by substrings separated 
  # by spaces within the strings instead of by the whole string elements
  for COMMAND in "${COMMANDS[@]}"; do

    # run it
    o=$($COMMAND)
    
    # if the command output exists, and is not an empty string, output it
    if [ o ] && [ ! "$o" == "" ]; then
      echo -e "\n======================" >> "${OUTPUTPATH}"
      echo -e "======================" >> "${OUTPUTPATH}"
      echo -e "=== shell: ${COMMAND}" >> "${OUTPUTPATH}"
      echo -e "=======" >> "${OUTPUTPATH}"
      echo -e "${o}" >> "${OUTPUTPATH}"
    fi

  done

  echo -e "\n============================================" >> "${OUTPUTPATH}"
  echo -e "= End System Information Report" >> "${OUTPUTPATH}"
  echo -e "============================================\n" >> "${OUTPUTPATH}"

fi


# use curl to upload the log data file to the server
CURLRESP=$(curl -F "host=${HOST}" -F "date=${DATE_S}" -F "log=@${OUTPUTPATH}" "${OUTPUTURL}")


# check the json response for 201 status text
if [[ "$CURLRESP" == *"\"status\":\"201\""* ]]; then
  # if OUTPUTDIR == /var/tmp/ then the file should be deleted on success
  # otherwise leave it in place
  if [[ "$OUTPUTDIR" == "/var/tmp/"  ]]; then
    # perform cleanup
    rm "${OUTPUTPATH}"
  fi

else
  echo -e "\nError:"
  echo -e "\n${CURLRESP}"
  # exit fail
  exit 1
fi


# exit success
exit 0

