dd if=/dev/zero of=masterfile bs=1 count=10000000
cd test_data

split -b 1024 -a 10 ../masterfile
cd ..

cp test_data 100 -fr
cp test_data 300 -fr
cp test_data 500 -fr
cp test_data 1000 -fr