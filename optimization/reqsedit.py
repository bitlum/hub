row_array = []

ignore_keys = ['grpc==', 'protobuf_to_dict==']

with open('requirements.txt') as file:
    for line in file:
        row_array.append(line)
        for i in range(len(ignore_keys)):
            if row_array[-1][:len(ignore_keys[i])] == ignore_keys[i]:
                row_array.pop()

with open('requirements.txt', 'w') as file:
    for row in row_array:
        file.write(row)
