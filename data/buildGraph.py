import matplotlib.pyplot as plt
import csv
import os 
import numpy as np
import datetime
import timple

tmpl = timple.Timple()
tmpl.enable()

dir_path = os.path.dirname(os.path.realpath(__file__))

print(dir_path)


client_range = range(20, 201, 20)

def get_average_from_csv(location):
    average_times = []
    for numClients in client_range:
        with open(location+str(numClients)+'_clients.csv', newline='') as csvfile:
            spamreader = csv.reader(csvfile, delimiter=',', quotechar='|')

            for row in spamreader:
                av_time = sum([int(x) for x in row if x != ""])/len(row)
                average_times.append(datetime.timedelta(microseconds=int(av_time/1000)))
    return np.array(average_times)

xpoints = [x for x in client_range]

average_times_TCP = get_average_from_csv(dir_path + "\\messageSize\\TCP\\")
average_times_QUIC = get_average_from_csv(dir_path + "\\messageSize\\QUIC\\")
average_times_UDP = get_average_from_csv(dir_path + "\\messageSize\\UDP\\")



plt.plot(xpoints, average_times_TCP, 'o')
# plt.plot(xpoints, average_times_QUIC)
plt.plot(xpoints, average_times_UDP, 'x')
plt.show()