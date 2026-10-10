local b = import 'builders.libsonnet';
local server(name) = 'windows_smb_server_shares_' + name + '{job=~"$job", instance="$instance"}';
local serverRate(name) = 'rate(' + server(name) + '[$__rate_interval])';
local client(name) = 'windows_smbclient_' + name + '{job=~"$job", instance="$instance"}';
local clientRate(name) = 'rate(' + client(name) + '[$__rate_interval])';
local serverRequirement = ' Requires the smb collector, which is not enabled by default.';
local clientRequirement = ' Requires the smbclient collector, which is not enabled by default.';
// Shares this host serves, and \\server\share paths this host connects to.
local share = '{{share}}';
local path = '\\\\{{server}}\\{{share}}';

function(on)
  b.tab(
    'SMB',
    b.rows([
      b.summary([
        if on('smb') then 220,
        if on('smb') then 221,
        if on('smb') then 222,
        if on('smbclient') then 223,
        if on('smbclient') then 224,
        if on('smbclient') then 225,
      ]),
      if on('smb') then b.row('Server shares', b.flow([[226, 12, 8], [227, 12, 8], [228, 8, 8], [229, 8, 8], [230, 8, 8]])),
      if on('smbclient') then b.row('Client shares', b.flow([[231, 12, 8], [232, 12, 8], [233, 12, 8], [234, 12, 8], [235, 12, 8], [236, 12, 8]])),
    ]),
    [
      b.stat(220, 'Open files', 'Files open on the shares of this SMB server.' + serverRequirement, 'sum(' + server('current_open_file_count') + ')'),
      b.stat(221, 'Connections', 'Client connections (tree connects) to the shares of this SMB server.' + serverRequirement, 'sum(' + server('tree_connect_count') + ')'),
      b.stat(222, 'Server traffic', 'Bytes sent and received per second by the shares of this SMB server.' + serverRequirement, 'sum(' + serverRate('received_bytes_total') + ') + sum(' + serverRate('sent_bytes_total') + ')', 'Bps'),
      b.stat(223, 'Client traffic', 'Bytes read and written per second by this host on remote SMB shares.' + clientRequirement, 'sum(' + clientRate('data_bytes_total') + ')', 'Bps'),
      b.stat(224, 'Client latency', 'Average time per request of this host on remote SMB shares. 0 when there are no requests.' + clientRequirement, '(sum(' + clientRate('request_seconds_total') + ') / (sum(' + clientRate('requests_total') + ') > 0)) or (0 * sum(' + clientRate('requests_total') + '))', 's', steps=b.levels(0.02, 0.05)),
      b.stat(225, 'Credit stalls', 'Requests per second delayed because the server granted too few SMB credits. Sustained stalls limit throughput.' + clientRequirement, 'sum(' + clientRate('stalls_total') + ')', 'ops', steps=[{ color: 'green', value: null }, { color: 'orange', value: 1 }]),

      b.graph(226, 'Share traffic', 'Bytes sent and received per second per share. Received bytes are drawn below the axis.' + serverRequirement, 'Bps', [
        [serverRate('sent_bytes_total'), share + ' sent'],
        [serverRate('received_bytes_total'), share + ' received'],
      ], below='.* received$'),
      b.graph(227, 'Share requests', 'Read, write and metadata requests per second per share.' + serverRequirement, 'ops', [
        [serverRate('read_requests_count_total'), share + ' read'],
        [serverRate('write_requests_count_total'), share + ' write'],
        [serverRate('metadata_requests_count_total'), share + ' metadata'],
      ]),
      b.graph(228, 'Open files', 'Files open per share.' + serverRequirement, 'short', [[server('current_open_file_count'), share]]),
      b.graph(229, 'Connections', 'Client connections (tree connects) per share.' + serverRequirement, 'short', [[server('tree_connect_count'), share]]),
      b.graph(230, 'Files opened', 'Files opened per second per share.' + serverRequirement, 'ops', [[serverRate('files_opened_count_total'), share]]),

      b.graph(231, 'Client traffic', 'Bytes read and written per second per remote share. Reads are drawn below the axis.' + clientRequirement, 'Bps', [
        [clientRate('write_bytes_total'), path + ' write'],
        [clientRate('read_bytes_total'), path + ' read'],
      ], below='.* read$'),
      b.graph(232, 'Client requests', 'Read and write requests per second per remote share. Reads are drawn below the axis.' + clientRequirement, 'ops', [
        [clientRate('write_requests_total'), path + ' write'],
        [clientRate('read_requests_total'), path + ' read'],
      ], below='.* read$'),
      b.graph(233, 'Client latency', 'Average time per read and write request per remote share. Reads are drawn below the axis.' + clientRequirement, 's', [
        [clientRate('write_seconds_total') + ' / (' + clientRate('write_requests_total') + ' > 0)', path + ' write'],
        [clientRate('read_seconds_total') + ' / (' + clientRate('read_requests_total') + ' > 0)', path + ' read'],
      ], below='.* read$'),
      b.graph(234, 'Client queue length', 'Requests outstanding per remote share at the time of the scrape.' + clientRequirement, 'short', [[client('requests_queued'), path]]),
      b.graph(235, 'Credit stalls', 'Requests per second per remote share delayed because the server granted too few SMB credits.' + clientRequirement, 'ops', [[clientRate('stalls_total'), path]]),
      b.graph(236, 'Metadata requests', 'Metadata requests per second per remote share, such as opening files and listing directories.' + clientRequirement, 'ops', [[clientRate('metadata_requests_total'), path]]),
    ]
  )
