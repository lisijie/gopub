FROM alpine
MAINTAINER lsj86@qq.com

ENV DOWN_URL https://github.com/lisijie/gopub/releases/download/v2.0.1/gopub-v2.0.1-linux-amd64.tar.gz

RUN apk --no-cache add wget \
	&& mkdir gopub && cd gopub \
	&& wget --no-check-certificate -O gopub.tar.gz ${DOWN_URL} \
	&& tar fxz gopub.tar.gz && rm -f gopub.tar.gz

CMD ["/gopub/gopub"]