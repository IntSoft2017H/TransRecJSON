#!/usr/bin/env python

# Require pandas >= 0.21.0

import argparse
import logging
import pandas as pd
import sys

def line_count(filepath):
  """
  Count lines contained in a file
  License (CC BY-SA 2.5): Michael Bacon https://stackoverflow.com/a/27518377/5989200
  
  @param filepath file path string
  @return an integer of lines
  """
  f = open(filepath, 'rb')
  lines = 0
  buf_size = 1024 * 1024
  read_f = f.raw.read
  
  buf = read_f(buf_size)
  while buf:
    lines += buf.count(b'\n')
    buf = read_f(buf_size)
    
  return lines

def print_json_gz(filepath, json_uName, json_iName, json_value, json_voteTime):
  """
  Print data in a JSON dataset in the space-separeted text format used by TransRec
  
  @param filepath file path string
  @param json_uName user name string
  @param json_iName item name string
  @param json_value rating value integer
  @param json_voteTime rated time integer (e.g. unix time)
  @return void
  """
  whole_rows = line_count(filepath)
  row_size = 10000
  df_reader = pd.read_json(filepath,
                           lines=True,
                           chunksize=row_size,
                           compression='infer')
  row_nums = 0
  for chunk in df_reader:
    for _, row in chunk.iterrows():
      print(row[json_uName],
            row[json_iName],
            row[json_value],
            row[json_voteTime])
    row_nums += len(chunk)
    logging.info('completed %d rows (%d%%)' % (row_nums, row_nums * 100 / whole_rows))

def parse_args():
  """
  A parser of argument of this program
  
  @return a namespace object of arguments
  """
  epilog_text = '''example:
  %(prog)s reviews_Office_Products_5.json.gz | gzip > reviews_Office_Products_5.txt.gz
  %(prog)s -d googlelocal reviews.clean.json.gz > reviews.clean.txt
'''
  parser = argparse.ArgumentParser(description='Convert reviews.json.gz to reviews.txt',
                                   epilog=epilog_text,
                                   formatter_class=argparse.RawDescriptionHelpFormatter)
  parser.add_argument('-d', '--dataset',
                      default='amazon',
                      choices=['amazon', 'googlelocal'],
                      help='kind of dataset (default: amazon)')
  parser.add_argument('filepath',
                      help='input file (gzipped json)')
  return parser.parse_args()
    
if __name__ == '__main__':
  args = parse_args()

  logging.basicConfig(stream=sys.stderr,
                      level=logging.DEBUG,
                      format='%(asctime)s %(levelname)-8s %(message)s',
                      datefmt='%Y-%m-%d %H:%M:%S')
  if args.dataset == 'amazon':
    print_json_gz(filepath=args.filepath,
                  json_uName = 'reviewerID',
                  json_iName = 'asin',
                  json_value = 'overall',
                  json_voteTime = 'unixReviewTime')
  elif args.dataset == 'googlelocal':    
    print_json_gz(filepath=args.filepath,
                  json_uName = 'gPlusUserId',
                  json_iName = 'gPlusPlaceId',
                  json_value = 'rating',
                  json_voteTime = 'unixReviewTime')
  else:
    raise ValueError('Unknown kind of dataset: %s' % args.dataset)
