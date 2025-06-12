#!/bin/bash
for file in ./testing/*.js; do
	echo "Running $file..."
	k6 run "$file"
done

# for i in $(seq 1 100); do
# 	curl -i http://localhost:8080/JCawU4I
# done
