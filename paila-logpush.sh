#!/bin/bash

# allows the last command in a pipeline to run in the current shell,
# thus allowing variable changes within a while loop to persist
shopt -s lastpipe # Enable lastpipe option

# curl http file push location
OUTPUTURL="http://localhost:8181/uploadlog"

# todo : check for env variable to override OUTPUTURL
# todo : check for bash argv to override OUTPUTURL
# todo : --help




# setup some date objects for filtering and naming
DATE_S=$(date --date="yesterday"  +"%Y-%m-%d")
DATE_Y=$(date --date="yesterday"  +"%Y-%m-%d 00:00:00")
DATE_T=$(date --date="today"  +"%Y-%m-%d 00:00:00")

# regex pattern match for date string existance
DATE_P="[0-9]{4}-[0-9]{2}-[0-9]{2}"

# set the log dir and scan it
LOG_DIR="/var/log"
LOG_FILES=$(find $LOG_DIR -name "*.log" -mtime -3)

# determine the hostname
HOST=$(cat /etc/hostname)

# set output path for generated report file
OUTPUTPATH="${HOST}--${DATE_S}.report.txt"

# init logsfound to 0, change to 1 if anything comes up
# so the script will either gather further system information
# or generate a "good health" report
LOGSFOUND=0

# a generic header for the logged issues
echo -e "\n============================================" > "${OUTPUTPATH}"
echo -e "= Begin Logged Issues Report" >> "${OUTPUTPATH}"
echo -e "= Host: ${HOST}" >> "${OUTPUTPATH}"
echo -e "= Date: ${DATE_S}" >> "${OUTPUTPATH}"
echo -e "============================================" >> "${OUTPUTPATH}"

for LOGFILE in $LOG_FILES; do

  # get 1 possible error back from the file, this content will
  # be used to determine whether to dig deeper, and to check if
  # a date format is available to filter by
  grep -Eiw -m 1 "warning|error|critical|alert|fatal" $LOGFILE | while read -r checkline; do
    # check if a date filter can be used
    if [[ "$checkline" =~ $DATE_P ]]; then
      # date is available for filtering, get a count of the errors with the new filter applied
      LOGCOUNT=$(grep -Eiw "warning|error|critical|alert|fatal" $LOGFILE | grep -c "$DATE_S")
      # todo : should be > 0
      if [ "$LOGCOUNT" -gt 1 ]; then
        LOGSFOUND=1
        # date filtered log entries exist
        echo -e "\n======================" >> "${OUTPUTPATH}"
        echo -e "======================" >> "${OUTPUTPATH}"
        echo -e "=== Log: ${LOGFILE}" >> "${OUTPUTPATH}"
        echo -e "=======" >> "${OUTPUTPATH}"
        # do the output search
        grep -Eiw "warning|error|critical|alert|fatal" $LOGFILE | grep "$DATE_S" | while read -r line; do
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
      grep -Eiw "warning|error|critical|alert|fatal" $LOGFILE | while read -r line; do
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

# check if there were any logs found
if [ "$LOGSFOUND" -eq 1 ]; then
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

  # double quotes around "${array[@]}" are really important. Without them, the for loop 
  # will break up the array by substrings separated by spaces within the strings instead
  # of by the whole string elements
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


CURLRESP=$(curl -F "host=${HOST}" -F "date=${DATE_S}" -F "log=@${OUTPUTPATH}" "${OUTPUTURL}")


if [[ "$CURLRESP" == *"\"status\":\"201\""* ]]; then
  # perform cleanup
  echo -e "\n response is 201"
  rm "${OUTPUTPATH}"
else
  echo -e "\n${CURLRESP}"
  exit 1
fi

exit 0
