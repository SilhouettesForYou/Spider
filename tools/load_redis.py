import getopt
import redis
import sys
import numpy as np


def get_all_keys(r):
    keys = r.keys()
    return keys

def is_key_exist(r, key):
    keys = get_all_keys(r)
    keys = [k.decode() for k in keys]
    return key in keys


def get_values(r, key):
    values = []
    if not is_key_exist(r, key):
        print(key + ' is not int redis.')
        sys.exit(2)
    for value in r.smembers(key):
        values.append(value.decode())
    return np.array(values)


def load_and_save(r, filename, key):
    with np.load('./tools/' + filename) as values:
        if not is_key_exist(r, key):
            print(key + ' is not int redis.')
            sys.exit(2)
        for value in values:
            r.sadd(key, value)


def save_all(r):
    keys = [k.decode() for k in get_all_keys(r)]
    for key in keys:
        values = get_values(r, key)
        np.save('./tools/' + key, values)


def load_and_save_all(r):
    for file in os.listdir('./tools/'):
        if file.endswith('.npy'):
            load_and_save(r, './tools/' + file, file[:-4])


def usage(keys):
    print('usage of load data from redis')
    if len(keys) > 0:
        print('exist keys:')
    for key in keys:
        print('---' + key.decode())
    print('-h, --help    : print help message.')
    print('-k, --key     : the key of datas.')
    print('-s, --save    : save datas.')
    print('-l, --load    : load datas.')
    print('-u, --upload  : save all sets\' values to (.npy)')
    print('-d, --download: load all values from (.npy) and save to redis')


def parse_arg(argv, r):
    args = argv[1:]
    try:
        opts, args = getopt.getopt(args, 'hk:s:l:ud', ['help', 'key=', 'save=', 'laod=', 'upload', 'download'])
    except getopt.GetoptError as err:
        print(err)
        usage(get_all_keys(r))
        sys.exit(2)
    values = []
    key = ''
    for o, a in opts:
        if o in ('-h', '--help'):
            usage(get_all_keys(r))
            sys.exit(1)
        elif o in ('-k', '--key'):
            key = a
            values = get_values(r, a)
        elif o in ('-s', '--save'):
            if len(a) is 0:
                print('input file name after s or save.')
                exit(2)
            np.save('./tools/' + a, values)
        elif o in ('-l', '--load'):
            load_and_save(r, a, key)
        elif o in ('-u', '--upload'):
            save_all(r)
        elif o in ('-d', '--download'):
            load_and_save_all(r)


def main(argv):
    keys = []

    pool = redis.ConnectionPool(host='127.0.0.1', port=6379, db=0)
    r = redis.StrictRedis(connection_pool=pool)

    parse_arg(argv, r)


if __name__ == "__main__":
    main(sys.argv)