import test_random_sample
from smartsample import *

# number, mean, minimum, maximum, stdev
value = test_random_sample.calc(1000, 15, 10, 20, 1)

# list, prob_cut
sample = SmartSample(value, 0.4)
print(sample)

# print('value:')
# for i in range(observNumber):
#     print('{:.3f}'.format(value[i]), end=' ')



