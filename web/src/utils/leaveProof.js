export async function downloadLeaveProof(client, url, token, id) {
  const { data, headers } = await client.get(url, {
    headers: token ? { Authorization: `Bearer ${token}` } : undefined,
    responseType: 'blob'
  })
  const disposition = headers['content-disposition'] || ''
  const encodedName = disposition.match(/filename\*=utf-8''([^;]+)/i)?.[1]
  const plainName = disposition.match(/filename=(?:"((?:[^"\\]|\\.)*)"|([^;]+))/i)
  const extensions = { 'application/pdf': 'pdf', 'image/png': 'png', 'image/jpeg': 'jpg' }
  const fallback = `leave-proof-${id}.${extensions[data.type] || 'bin'}`
  const filename = encodedName
    ? decodeURIComponent(encodedName)
    : plainName?.[1]?.replace(/\\(.)/g, '$1') || plainName?.[2] || fallback

  const objectURL = URL.createObjectURL(data)
  const link = document.createElement('a')
  link.href = objectURL
  link.download = filename
  link.click()
  setTimeout(() => URL.revokeObjectURL(objectURL), 1000)
}
