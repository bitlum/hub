IMAGE:=hub:0.1

test:
	echo $(shell whoami)

build:
	docker build -t $(IMAGE) .

run:
	if [ ! -d log ] ; then mkdir log; fi
	docker run -ti \
		--mount type=bind,src=$(shell pwd)/log,dst=/optimization/optimizer/outlet \
		$(IMAGE)

push:
	docker push $(IMAGE)
