#!/bin/bash

output=$1
echo "Output file: $output"
if [ -f "$output" ]; then
    rm "$output"
fi

#grep -i "$text" ./result.txt >$output

# funcs
split() {
    line=$1
    separator=$2
    pos=$3
    IFS="$separator" read -a array <<< $line
    if [ $pos == "0" ]
    then
        echo ${array[0]}
    elif [ $pos == "1" ]
    then
        echo ${array[1]}
    elif [ $pos == "2" ]
    then
        echo ${array[2]}
    elif [ $pos == "3" ]
    then
        echo ${array[3]}
    elif [ $pos == "4" ]
    then
        echo ${array[4]}
    else
        echo ${array[*]}
    fi
}

#retval=$( split " liveobjs - 2698" "-" "1" )
#echo $retval

# read header
while read -r line; do
  data=${line:18}
  #echo $data
  IFS=',' read -a array <<< $data
  col0=$( split "${array[0]}" "-" "0" )
  col1=$( split "${array[1]}" "-" "0" )
  col2=$( split "${array[2]}" "-" "0" )
  col3=$( split "${array[3]}" "-" "0" )
  echo $col0, $col1, $col2, $col3 >  $output
  break
done < "./result.txt"

#tail  $output

# read data
while read -r line; do
  data=${line:18}
  #echo $data
  IFS=',' read -a array <<< $data
  col0=${array[0]:7}
  col1=$( split "${array[1]}" "-" "1" )
  col2=$( split "${array[2]}" "-" "1" )
  col3=$( split "${array[3]}" "-" "1" )
  echo $col0, $col1, $col2, $col3 >>  $output
done < "./result.txt"

#tail  $output
