#!/bin/bash

input="./mem.log"
echo "Analyze from file $input";

if [ ! -f $input ]; then
    echo "file not found: $input"
    exit
fi

# load pattern file
pattern="Memory Statistics"

# grep result
grep -i "$pattern" $input > ./result.txt
echo "output result into file ./result.txt"

./extravalue.sh ./data.csv