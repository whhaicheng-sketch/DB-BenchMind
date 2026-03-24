function normalizeText(value) {
  return String(value || '').trim().toLowerCase()
}

function hasTestWord(value) {
  const normalized = normalizeText(value)
  if (!normalized) return false
  return /(?:^|[\s:_-])test(?:$|[\s:_-])/.test(normalized)
}

export function isTestTemplate(template = {}) {
  if (normalizeText(template.profile_type) === 'test') {
    return true
  }

  if (Array.isArray(template.tags) && template.tags.some((tag) => normalizeText(tag) === 'test')) {
    return true
  }

  return [template.name, template.id, template.code, template.key].some((value) => hasTestWord(value))
}

export function getPreferredTemplateId(templates = [], { fallbackTemplateId = '' } = {}) {
  const availableTemplates = Array.isArray(templates) ? templates : []
  const preferredTemplate = availableTemplates.find((template) => isTestTemplate(template))

  if (preferredTemplate?.id) {
    return preferredTemplate.id
  }

  return availableTemplates.some((template) => template.id === fallbackTemplateId) ? fallbackTemplateId : ''
}
