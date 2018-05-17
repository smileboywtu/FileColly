dd if=/dev/zero of=masterfile bs=1 count=10000000
cd test_data
split -b 20 -a 10 ../masterfile


