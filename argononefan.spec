Name:     argononefan
Version:  0.11
Release:  %autorelease
Summary:  ArgonOne fan control daemon and cli tools
License:  MIT
URL:      https://github.com/mwmahlberg/argononefan/
Source0:  https://github.com/mwmahlberg/argononefan/archive/refs/heads/master.zip
BuildRequires: golang
%global debug_package %{nil}
%description
ArgonOne fan control daemon and cli tools.

%prep
echo "Prep"
%autosetup -S git -n %{name}-master

%build -p argononefan-master
go build -ldflags=-linkmode=external ./cmd/setfan
go build -ldflags=-linkmode=external ./cmd/readtemp
go build -ldflags=-linkmode=external ./cmd/adjustfan

%install
rm -rf $RPM_BUILD_ROOT
install -D -m 0640 deploy/adjustfan.service $RPM_BUILD_ROOT/lib/systemd/system/adjustfan.service
install -D -m 0750 setfan $RPM_BUILD_ROOT/%{_sbindir}/setfan
install -D -m 0755 readtemp $RPM_BUILD_ROOT/%{_sbindir}/readtemp
install -D -m 0750 adjustfan $RPM_BUILD_ROOT/%{_sbindir}/adjustfan
install -D -m 0640 cmd/adjustfan/adjustfan.json $RPM_BUILD_ROOT/%{_sysconfdir}/argonone/adjustfan.json

%files
/lib/systemd/system/adjustfan.service
%{_sbindir}/setfan
%{_sbindir}/readtemp
%{_sbindir}/adjustfan

%config
%{_sysconfdir}/argonone/adjustfan.json

%changelog
%autochangelog