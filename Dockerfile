FROM python:3.6
MAINTAINER andrey.u@gmail.com

RUN apt-get update
RUN apt-get install -y screen

COPY manager/docker/hub/hub /usr/local/bin/manager
COPY optimization/ /optimization

RUN pip install -r /optimization/requirements.txt
RUN pip install numpy
RUN pip install git+https://github.com/oblalex/gnuplot.py-py3k.git

RUN if [ -d /optimization/optimizer/outlet ] ; then rm -rf /optimization/optimizer/outlet ; fi

# patch configs
RUN sed -i 's/python3/python/g' /optimization/cycle_run.sh 
RUN sed -i 's/"make_drawing": true/"make_drawing": false/' /optimization/optimizer/routermgt_inlet.json
RUN sed -i 's/"output_period": .$/"output_period": 15000/' /optimization/optimizer/routermgt_inlet.json

CMD bash -c "cd /optimization; ./cycle_run.sh |& tee /optimization/optimizer/outlet/std.log"
