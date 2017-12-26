# Convert JSON for TransRec

A script `json2txt.py` converts a JSON file into a text file which can be used as an input for TransRec.

## Usage

For Amazon datasets,

```sh
python3 ./json2txt.py reviews_Office_Products_5.json.gz | gzip > reviews_Office_Products_5.txt.gz
```

For other datasets, use `-d` option.

```sh
python3 ./json2txt.py -d googlelocal reviews.clean.json.gz | gzip > reviews.clean.txt
```

# Datasets for TransRec

Paper: <http://cseweb.ucsd.edu/~jmcauley/pdfs/recsys17.pdf>

## Compilation

Code: <https://drive.google.com/file/d/0B9Ck8jw-TZUEVmdROWZKTy1fcEE/view?usp=sharing>

### `sequential_rec`

1. Edit `src/main.cpp` to use `go_TransRec`.
2. `make clean`
3. `make`

You can use the result executable `train` as the following:

```sh
./train reviews_Automotive.txt.gz 5 5 10 0.1 0.1 0.01 0 10000 my_model_path_blah_blah
```

## Amazon

Data: <http://jmcauley.ucsd.edu/data/amazon/>

We can use `json2txt.py` to convert Amazon's 5-core json datasets to txt.

## Google Local

Data: <http://jmcauley.ucsd.edu/data/googlelocal/googlelocal.tar.gz>

We can use `json2txt.py` to convert json to txt.

## Epinions

Data: <http://jmcauley.ucsd.edu/data/epinions/>

## Foursquare

Data: <https://archive.org/details/201309_foursquare_dataset_umn>

## Flixter

Data: <strike>http://www.cs.ubc.ca/~jamalim/datasets/</strike> --> <http://socialcomputing.asu.edu/datasets/Flixster>
