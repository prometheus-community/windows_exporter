// manifestJson renders a value as JSON indented by four spaces, as .editorconfig requires for JSON
// files. Unlike std.manifestJsonEx, it writes empty arrays and objects as [] and {}.
// It concatenates strings instead of using std.format, which runs out of stack frames on large output.
local manifestJson(value, indent='') =
  local inner = indent + '    ';
  if std.isArray(value) then
    if std.length(value) == 0 then '[]'
    else '[\n' + std.join(',\n', [inner + manifestJson(v, inner) for v in value]) + '\n' + indent + ']'
  else if std.isObject(value) then
    local fields = std.objectFields(value);
    if std.length(fields) == 0 then '{}'
    else
      '{\n'
      + std.join(',\n', [inner + std.escapeStringJson(f) + ': ' + manifestJson(value[f], inner) for f in fields])
      + '\n' + indent + '}'
  else std.manifestJson(value);

// Render with jsonnet -S, which appends the final newline.
function(value) manifestJson(value)
