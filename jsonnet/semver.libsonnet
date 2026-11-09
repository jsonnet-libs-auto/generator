{
  // parse('1.2.3') == [1, 2, 3]; a leading 'v' is ignored.
  parse(v)::
    local s = if std.startsWith(v, 'v') then v[1:] else v;
    [std.parseInt(p) for p in std.split(s, '.')],

  // atLeast('1.27.4', '1.27') == true; missing parts of min count as 0.
  atLeast(v, min)::
    local a = self.parse(v), b = self.parse(min);
    local cmp(i) =
      if i >= std.length(b) then true
      else
        local x = if i < std.length(a) then a[i] else 0;
        if x != b[i] then x > b[i] else cmp(i + 1);
    cmp(0),
}
