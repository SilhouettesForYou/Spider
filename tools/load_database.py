import idb


with idb.from_file('D:\\Mysql-8.0.18-winx64\\data\\crawl\\newtimes.ibd') as db:
    api = idb.IDAPython(db)
    for ea in api.idautils.Functions():
        print('%x: %s' % (ea, api.GetFunctionName(ea)))