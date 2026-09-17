-- Rewrites relative .md link targets to .html in rendered pages.
function Link(el)
  if not el.target:match("^%a[%w+.-]*:") then
    el.target = el.target:gsub("%.md$", ".html"):gsub("%.md#", ".html#")
  end
  return el
end
