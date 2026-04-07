# This project uses the LITNET-2020 dataset to verify the comparison algorithm.
# More information about the dataset can be found at: https://www.mdpi.com/2079-9292/9/5/800

url="https://github.com/Grigaliunas/electronics9050800/raw/refs/heads/main/dataset/ALLinONE.zip?download="
data_archive="LITNET-2020.zip"
data_file="LITNET-2020.csv"
sample_size=1000
sample_file="s$sample_size-$data_file"

# sample random seed:
# command used to generate seed: od -An -N32 -tx1 /dev/urandom | tr -d " \n" && echo ""
seed=30b14cf307e68cc40f6415ccd6adae4e20c319e2b3ef3f798977f83225afee43

# Download dataset
wget $url -O $data_archive
unzip $data_archive -d data
mv data/allFlows.csv data/$data_file

# Sample data
data_size=$(wc -l data/$data_file | awk '{ print $1 }')
data_sample=$(head -n 1 data/$data_file && tail -n +2 data/$data_file | shuf --random-source=<(openssl enc -aes-256-ctr -pass pass:"$seed" -nosalt </dev/zero 2>/dev/null) -n $sample_size)
echo $data_sample > data/$sample_file

# Filter data and convert IPv4 addresses to one-byte hash variants.
python3 <<EOF
import csv
import hashlib

cols = ["te_year", "te_month", "te_day", "te_hour", "te_min", "te_second", "td", "sa", "da", "attack_t", "attack_a"]

with open("data/$sample_file", newline="") as infile, open("data/f$sample_file", "w", newline="") as outfile:
    reader = csv.DictReader(infile)
    writer = csv.DictWriter(outfile, fieldnames=cols)

    writer.writeheader()
    for row in reader:
        row["sa"] = hashlib.sha256(row["sa"].encode()).digest()[0]
        row["da"] = hashlib.sha256(row["da"].encode()).digest()[0]
        writer.writerow({c: row[c] for c in cols})
EOF

# Fetch Cyber Threat Intelligence
python3 <<EOF
import csv

with open("data/f$sample_file", newline="") as infile, open("data/threats-f$sample_file", "w", newline="") as outfile:
    beneign = []
    threat = []
    unclassified = []

    reader = csv.DictReader(infile)
    for row in reader:
        if row["attack_a"] == "1":
            threat.append(row["sa"])
        else:
            unclassified.extend([row["sa"], row["da"]])
    
    beneign = list(set(unclassified) - set(threat))
    
    writer = csv.DictWriter(outfile, fieldnames=["a"])
    writer.writeheader()
    for row in threat:
        writer.writerow({"a": row})

    print(len(unclassified))
    print(len(threat))
    print(len(beneign))
EOF