from samplegen import *
from smartsample import *

value = generate_sample(number=1000, mean=10, stdev=2)

# list, prob_cut
sample = SmartSample(value, 0.4)
print(sample)

# print('value:')
# for i in range(observNumber):
#     print('{:.3f}'.format(value[i]), end=' ')
