const defaultErrorMessage = '故事暂时无法加载，请稍后再试。'

async function request(path) {
  let response
  try {
    response = await fetch(path, {
      headers: { Accept: 'application/json' },
    })
  } catch {
    throw new Error(defaultErrorMessage)
  }

  if (!response.ok) {
    const payload = await response.json().catch(() => ({}))
    throw new Error(payload.error || defaultErrorMessage)
  }

  return response.json()
}

export async function fetchNovel() {
  return request('/api/novel')
}

export async function fetchChapter(id) {
  return request(`/api/chapters/${id}`)
}
