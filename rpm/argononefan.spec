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
make

%install
rm -rf $RPM_BUILD_ROOT
install -D -m 0640 rpm/argononefan.service $RPM_BUILD_ROOT/lib/systemd/system/argononefan.service
install -D -m 0750 argononefan $RPM_BUILD_ROOT/%{_sbindir}/argononefan

%files
/lib/systemd/system/adjustfan.service
%{_sbindir}/setfan
%{_sbindir}/argononefand

%config
%{_sysconfdir}/argonone/adjustfan.json

%changelog
%autochangelog