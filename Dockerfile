ARG FEDORA_VERSION=39
FROM fedora:${FEDORA_VERSION}
RUN dnf install -y rpm-build rpmdevtools rpmautospec git \
  && rpmdev-setuptree
# ADD argononefan.spec /root/rpmbuild/SPECS/argononefan.spec
# RUN spectool -g -R rpmbuild/SPECS/argononefan.spec && \
#     rpmbuild -ba rpmbuild/SPECS/argononefan.spec