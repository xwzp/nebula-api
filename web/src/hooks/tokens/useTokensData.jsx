/*
Copyright (C) 2025 QuantumNous

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as
published by the Free Software Foundation, either version 3 of the
License, or (at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program. If not, see <https://www.gnu.org/licenses/>.

For commercial licensing, please contact support@quantumnous.com
*/

import { useState, useEffect, useRef } from 'react';
import { useTranslation } from 'react-i18next';
import { Modal } from '@douyinfe/semi-ui';
import {
  API,
  copy,
  showError,
  showSuccess,
  encodeToBase64,
} from '../../helpers';
import { ITEMS_PER_PAGE } from '../../constants';
import { useTableCompactMode } from '../common/useTableCompactMode';
import {
  fetchTokenKey as fetchTokenKeyById,
  getServerAddress,
} from '../../helpers/token';

export const useTokensData = (openFluentNotification, openCCSwitchModal) => {
  const { t } = useTranslation();

  // Basic state
  const [tokens, setTokens] = useState([]);
  const [loading, setLoading] = useState(true);
  const [activePage, setActivePage] = useState(1);
  const [tokenCount, setTokenCount] = useState(0);
  const [pageSize, setPageSize] = useState(ITEMS_PER_PAGE);
  const [searching, setSearching] = useState(false);
  const [searchMode, setSearchMode] = useState(false); // 是否处于搜索结果视图

  // Selection state
  const [selectedKeys, setSelectedKeys] = useState([]);

  // Edit state
  const [showEdit, setShowEdit] = useState(false);
  const [editingToken, setEditingToken] = useState({
    id: undefined,
  });

  // UI state
  const [compactMode, setCompactMode] = useTableCompactMode('tokens');
  const [showKeys, setShowKeys] = useState({});
  const [resolvedTokenKeys, setResolvedTokenKeys] = useState({});
  const [loadingTokenKeys, setLoadingTokenKeys] = useState({});
  const keyRequestsRef = useRef({});

  // Form state
  const [formApi, setFormApi] = useState(null);
  const formInitValues = {
    searchKeyword: '',
    searchToken: '',
  };

  // Get form values helper function
  const getFormValues = () => {
    const formValues = formApi ? formApi.getValues() : {};
    return {
      searchKeyword: formValues.searchKeyword || '',
      searchToken: formValues.searchToken || '',
    };
  };

  // Close edit modal
  const closeEdit = () => {
    setShowEdit(false);
    setTimeout(() => {
      setEditingToken({
        id: undefined,
      });
    }, 500);
  };

  // Sync page data from API response
  const syncPageData = (payload) => {
    setTokens(payload.items || []);
    setTokenCount(payload.total || 0);
    setActivePage(payload.page || 1);
    setPageSize(payload.page_size || pageSize);
    setShowKeys({});
  };

  // Load tokens function
  const loadTokens = async (page = 1, size = pageSize) => {
    setLoading(true);
    setSearchMode(false);
    const res = await API.get(`/api/token/?p=${page}&size=${size}`);
    const { success, message, data } = res.data;
    if (success) {
      syncPageData(data);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  // Refresh function
  const refresh = async (page = activePage) => {
    await loadTokens(page);
    setSelectedKeys([]);
  };

  // Copy text function
  const copyText = async (text) => {
    if (await copy(text)) {
      showSuccess(t('已复制到剪贴板！'));
    } else {
      Modal.error({
        title: t('无法复制到剪贴板，请手动复制'),
        content: text,
        size: 'large',
      });
    }
  };

  const fetchTokenKey = async (tokenOrId, options = {}) => {
    const { suppressError = false } = options;
    const tokenId =
      typeof tokenOrId === 'object' ? tokenOrId?.id : Number(tokenOrId);

    if (!tokenId) {
      const error = new Error(t('令牌不存在'));
      if (!suppressError) {
        showError(error.message);
      }
      throw error;
    }

    if (resolvedTokenKeys[tokenId]) {
      return resolvedTokenKeys[tokenId];
    }

    if (keyRequestsRef.current[tokenId]) {
      return keyRequestsRef.current[tokenId];
    }

    const request = (async () => {
      setLoadingTokenKeys((prev) => ({ ...prev, [tokenId]: true }));
      try {
        const fullKey = await fetchTokenKeyById(tokenId);
        setResolvedTokenKeys((prev) => ({ ...prev, [tokenId]: fullKey }));
        return fullKey;
      } catch (error) {
        const normalizedError = new Error(
          error?.message || t('获取令牌密钥失败'),
        );
        if (!suppressError) {
          showError(normalizedError.message);
        }
        throw normalizedError;
      } finally {
        delete keyRequestsRef.current[tokenId];
        setLoadingTokenKeys((prev) => {
          const next = { ...prev };
          delete next[tokenId];
          return next;
        });
      }
    })();

    keyRequestsRef.current[tokenId] = request;
    return request;
  };

  const toggleTokenVisibility = async (record) => {
    const tokenId = record?.id;
    if (!tokenId) {
      return;
    }

    if (showKeys[tokenId]) {
      setShowKeys((prev) => ({ ...prev, [tokenId]: false }));
      return;
    }

    const fullKey = await fetchTokenKey(record);
    if (fullKey) {
      setShowKeys((prev) => ({ ...prev, [tokenId]: true }));
    }
  };

  const copyTokenKey = async (record) => {
    const fullKey = await fetchTokenKey(record);
    await copyText(`sk-${fullKey}`);
  };

  const shellSingleQuote = (value) => {
    return "'" + String(value).replace(/'/g, "'\\''") + "'";
  };

  const buildPythonStdinCommand = (pythonScript, payload, marker, shell) => {
    if (shell === 'fish') {
      return `printf '%s\\n' ${shellSingleQuote(payload)} | python3 -c ${shellSingleQuote(pythonScript)}`;
    }
    return `python3 -c ${shellSingleQuote(pythonScript)} << '${marker}'\n${payload}\n${marker}`;
  };

  const buildAsciiSlug = (value, fallback) => {
    const slug = String(value || '')
      .normalize('NFKD')
      .replace(/[\u0300-\u036f]/g, '')
      .toLowerCase()
      .replace(/&/g, ' and ')
      .replace(/[^a-z0-9]+/g, '-')
      .replace(/^-+|-+$/g, '');
    return slug || fallback;
  };

  const buildOpenClawCommand = (record, fullKey, models, shell) => {
    const serverAddress = getServerAddress();
    const providerKey =
      'nebula-' + (record.name || 'default').replace(/\s+/g, '-');
    const providerConfig = {};
    providerConfig[providerKey] = {
      baseUrl: serverAddress + '/v1',
      apiKey: `sk-${fullKey}`,
      models: models || [],
    };
    const providerJson = JSON.stringify(providerConfig, null, 2);

    const pythonScript = `import json, os, sys

provider = json.loads(sys.stdin.read())
path = os.path.expanduser("~/.openclaw/openclaw.json")
try:
    with open(path) as f:
        existing = json.load(f)
except (FileNotFoundError, json.JSONDecodeError):
    existing = {}

existing.setdefault("models", {}).setdefault("providers", {}).update(provider)

pk = list(provider.keys())[0]
models = provider[pk].get("models", [])
dm = existing.setdefault("agents", {}).setdefault("defaults", {}).setdefault("models", {})
prefix = pk + "/"
for k in list(dm.keys()):
    if k.startswith(prefix):
        del dm[k]
for m in models:
    dm[prefix + m["id"]] = {}

os.makedirs(os.path.dirname(path), exist_ok=True)
with open(path, "w") as f:
    json.dump(existing, f, indent=2, ensure_ascii=False)

print("Done! Nebula provider configured at " + path)`;

    return buildPythonStdinCommand(
      pythonScript,
      providerJson,
      'OPENCLAW_JSON',
      shell,
    );
  };

  const buildHermesCommand = (record, fullKey, models, shell) => {
    const serverAddress = getServerAddress();
    const providerName =
      'nebula-' +
      buildAsciiSlug(
        record.name || 'default',
        `token-${record.id || 'default'}`,
      );
    const hermesModels = {};

    for (const model of models || []) {
      if (!model?.id) {
        continue;
      }
      hermesModels[model.id] = {
        context_length: Number(model.contextWindow) || 128000,
        max_tokens: Number(model.maxTokens) || 4096,
      };
    }

    const providerConfig = {
      name: providerName,
      base_url: serverAddress + '/v1',
      api_key: `sk-${fullKey}`,
      api_mode: 'codex_responses',
      model: 'gpt-5.5',
      models: hermesModels,
    };
    const providerJson = JSON.stringify(providerConfig, null, 2);

    const pythonScript = `import json, os, re, sys
from pathlib import Path

provider = json.loads(sys.stdin.read())
home = os.environ.get("HERMES_HOME", "").strip()
path = (Path(home).expanduser() if home else Path.home() / ".hermes") / "config.yaml"

def yaml_scalar(value):
    text = str(value)
    if re.match(r"^[A-Za-z0-9_./@:+-]+$", text) and not text.startswith(("-", "?", ":", "@", "\`")):
        return text
    return json.dumps(text, ensure_ascii=False)

def yaml_key(value):
    text = str(value)
    if re.match(r"^[A-Za-z0-9_./@+-]+$", text):
        return text
    return json.dumps(text, ensure_ascii=False)

def parse_name(raw):
    text = raw.strip()
    if text.startswith('"'):
        try:
            return json.loads(text)
        except Exception:
            pass
    if text.startswith("'") and text.endswith("'"):
        return text[1:-1].replace("''", "'")
    return text

def build_entry(entry, indent=""):
    child = indent + "  "
    model_indent = child + "  "
    lines = [
        indent + "- name: " + yaml_scalar(entry["name"]),
        child + "base_url: " + yaml_scalar(entry["base_url"]),
        child + "api_key: " + yaml_scalar(entry["api_key"]),
        child + "api_mode: " + yaml_scalar(entry["api_mode"]),
        child + "model: " + yaml_scalar(entry["model"]),
    ]
    models = entry.get("models") or {}
    if models:
        lines.append(child + "models:")
        for model_name, model_cfg in models.items():
            context_length = int(model_cfg.get("context_length") or 128000)
            max_tokens = int(model_cfg.get("max_tokens") or 4096)
            lines.append(model_indent + yaml_key(model_name) + ":")
            lines.append(model_indent + "  context_length: " + str(context_length))
            lines.append(model_indent + "  max_tokens: " + str(max_tokens))
    else:
        lines.append(child + "models: {}")
    return lines

def is_top_level_key(line):
    return re.match(r"^[A-Za-z_][A-Za-z0-9_-]*\\s*:", line) is not None

try:
    text = path.read_text()
except FileNotFoundError:
    text = ""

lines = text.splitlines()
cp_index = None
for i, line in enumerate(lines):
    if re.match(r"^custom_providers\\s*:", line):
        cp_index = i
        break

if cp_index is None:
    new_text = text
    if new_text and not new_text.endswith("\\n"):
        new_text += "\\n"
    if new_text:
        new_text += "\\n"
    new_text += "custom_providers:\\n" + "\\n".join(build_entry(provider)) + "\\n"
else:
    if re.match(r"^custom_providers\\s*:\\s*\\[\\s*\\]\\s*(#.*)?$", lines[cp_index]):
        lines[cp_index] = "custom_providers:"

    end = len(lines)
    for i in range(cp_index + 1, len(lines)):
        line = lines[i]
        if line.strip() == "" or line.lstrip().startswith("#"):
            continue
        if is_top_level_key(line):
            end = i
            break

    indent = ""
    for line in lines[cp_index + 1:end]:
        m = re.match(r"^(\\s*)-\\s+", line)
        if m:
            indent = m.group(1)
            break

    block = lines[cp_index + 1:end]
    filtered = []
    i = 0
    while i < len(block):
        line = block[i]
        m = re.match(r"^" + re.escape(indent) + r"-\\s+name\\s*:\\s*(.*?)\\s*(?:#.*)?$", line)
        if m and parse_name(m.group(1)) == provider["name"]:
            i += 1
            while i < len(block):
                if re.match(r"^" + re.escape(indent) + r"-\\s+", block[i]):
                    break
                i += 1
            continue
        filtered.append(line)
        i += 1

    while filtered and filtered[-1].strip() == "":
        filtered.pop()

    entry_lines = build_entry(provider, indent)
    new_lines = lines[:cp_index + 1] + filtered + entry_lines + lines[end:]
    new_text = "\\n".join(new_lines) + "\\n"

path.parent.mkdir(parents=True, exist_ok=True)
path.write_text(new_text)
print("Done! Nebula provider configured at " + str(path))`;

    return buildPythonStdinCommand(
      pythonScript,
      providerJson,
      'HERMES_JSON',
      shell,
    );
  };

  // Generate provider config and copy shell command to clipboard
  // shell: 'bash' | 'fish'
  const onConfigureProvider = async (record, provider, shell = 'bash') => {
    if (provider !== 'openclaw' && provider !== 'hermes') return;
    try {
      const [fullKey, res] = await Promise.all([
        fetchTokenKey(record),
        API.get(`/api/token/${record.id}/openclaw-models`),
      ]);
      const { success, message, data } = res.data;
      if (!success) {
        showError(message || t('获取模型列表失败'));
        return;
      }

      const command =
        provider === 'hermes'
          ? buildHermesCommand(record, fullKey, data, shell)
          : buildOpenClawCommand(record, fullKey, data, shell);

      await copyText(command);
      showSuccess(
        provider === 'hermes'
          ? t('Hermes 配置命令已复制到剪贴板')
          : t('OpenClaw 配置命令已复制到剪贴板'),
      );
    } catch (error) {
      showError(error?.message || t('生成配置失败'));
    }
  };

  // Open link function for chat integrations
  const onOpenLink = async (type, url, record) => {
    const fullKey = await fetchTokenKey(record);
    if (url && url.startsWith('ccswitch')) {
      openCCSwitchModal(fullKey);
      return;
    }
    if (url && url.startsWith('fluent')) {
      openFluentNotification(fullKey);
      return;
    }
    const serverAddress = getServerAddress();
    if (url.includes('{cherryConfig}') === true) {
      let cherryConfig = {
        id: 'new-api',
        baseUrl: serverAddress,
        apiKey: `sk-${fullKey}`,
      };
      let encodedConfig = encodeURIComponent(
        encodeToBase64(JSON.stringify(cherryConfig)),
      );
      url = url.replaceAll('{cherryConfig}', encodedConfig);
    } else if (url.includes('{aionuiConfig}') === true) {
      let aionuiConfig = {
        platform: 'new-api',
        baseUrl: serverAddress,
        apiKey: `sk-${fullKey}`,
      };
      let encodedConfig = encodeURIComponent(
        encodeToBase64(JSON.stringify(aionuiConfig)),
      );
      url = url.replaceAll('{aionuiConfig}', encodedConfig);
    } else {
      let encodedServerAddress = encodeURIComponent(serverAddress);
      url = url.replaceAll('{address}', encodedServerAddress);
      url = url.replaceAll('{key}', `sk-${fullKey}`);
    }

    window.open(url, '_blank');
  };

  // Manage token function (delete, enable, disable)
  const manageToken = async (id, action, record) => {
    setLoading(true);
    let data = { id };
    let res;
    switch (action) {
      case 'delete':
        res = await API.delete(`/api/token/${id}/`);
        break;
      case 'enable':
        data.status = 1;
        res = await API.put('/api/token/?status_only=true', data);
        break;
      case 'disable':
        data.status = 2;
        res = await API.put('/api/token/?status_only=true', data);
        break;
    }
    const { success, message } = res.data;
    if (success) {
      showSuccess(t('操作成功完成！'));
      let token = res.data.data;
      let newTokens = [...tokens];
      if (action !== 'delete') {
        record.status = token.status;
      }
      setTokens(newTokens);
    } else {
      showError(message);
    }
    setLoading(false);
  };

  // Search tokens function
  const searchTokens = async (page = 1, size = pageSize) => {
    const normalizedPage = Number.isInteger(page) && page > 0 ? page : 1;
    const normalizedSize = Number.isInteger(size) && size > 0 ? size : pageSize;

    const { searchKeyword, searchToken } = getFormValues();
    if (searchKeyword === '' && searchToken === '') {
      setSearchMode(false);
      await loadTokens(1);
      return;
    }
    setSearching(true);
    const res = await API.get(
      `/api/token/search?keyword=${encodeURIComponent(searchKeyword)}&token=${encodeURIComponent(searchToken)}&p=${normalizedPage}&size=${normalizedSize}`,
    );
    const { success, message, data } = res.data;
    if (success) {
      setSearchMode(true);
      syncPageData(data);
    } else {
      showError(message);
    }
    setSearching(false);
  };

  // Sort tokens function
  const sortToken = (key) => {
    if (tokens.length === 0) return;
    setLoading(true);
    let sortedTokens = [...tokens];
    sortedTokens.sort((a, b) => {
      return ('' + a[key]).localeCompare(b[key]);
    });
    if (sortedTokens[0].id === tokens[0].id) {
      sortedTokens.reverse();
    }
    setTokens(sortedTokens);
    setLoading(false);
  };

  // Page handlers
  const handlePageChange = (page) => {
    if (searchMode) {
      searchTokens(page, pageSize).then();
    } else {
      loadTokens(page, pageSize).then();
    }
  };

  const handlePageSizeChange = async (size) => {
    setPageSize(size);
    if (searchMode) {
      await searchTokens(1, size);
    } else {
      await loadTokens(1, size);
    }
  };

  // Row selection handlers
  const rowSelection = {
    onSelect: (record, selected) => {},
    onSelectAll: (selected, selectedRows) => {},
    onChange: (selectedRowKeys, selectedRows) => {
      setSelectedKeys(selectedRows);
    },
  };

  // Handle row styling
  const handleRow = (record, index) => {
    if (record.status !== 1) {
      return {
        style: {
          background: 'var(--semi-color-disabled-border)',
        },
      };
    } else {
      return {};
    }
  };

  // Batch delete tokens
  const batchDeleteTokens = async () => {
    if (selectedKeys.length === 0) {
      showError(t('请先选择要删除的令牌！'));
      return;
    }
    setLoading(true);
    try {
      const ids = selectedKeys.map((token) => token.id);
      const res = await API.post('/api/token/batch', { ids });
      if (res?.data?.success) {
        const count = res.data.data || 0;
        showSuccess(t('已删除 {{count}} 个令牌！', { count }));
        await refresh();
        setTimeout(() => {
          if (tokens.length === 0 && activePage > 1) {
            refresh(activePage - 1);
          }
        }, 100);
      } else {
        showError(res?.data?.message || t('删除失败'));
      }
    } catch (error) {
      showError(error.message);
    } finally {
      setLoading(false);
    }
  };

  // Batch copy tokens
  const batchCopyTokens = async (copyType) => {
    if (selectedKeys.length === 0) {
      showError(t('请至少选择一个令牌！'));
      return;
    }
    try {
      const keys = await Promise.all(
        selectedKeys.map((token) =>
          fetchTokenKey(token, { suppressError: true }),
        ),
      );
      let content = '';
      for (let i = 0; i < selectedKeys.length; i++) {
        const fullKey = keys[i];
        if (copyType === 'name+key') {
          content += `${selectedKeys[i].name}    sk-${fullKey}\n`;
        } else {
          content += `sk-${fullKey}\n`;
        }
      }
      await copyText(content);
    } catch (error) {
      showError(error?.message || t('复制令牌失败'));
    }
  };

  // Initialize data
  useEffect(() => {
    loadTokens(1)
      .then()
      .catch((reason) => {
        showError(reason);
      });
  }, [pageSize]);

  return {
    // Basic state
    tokens,
    loading,
    activePage,
    tokenCount,
    pageSize,
    searching,

    // Selection state
    selectedKeys,
    setSelectedKeys,

    // Edit state
    showEdit,
    setShowEdit,
    editingToken,
    setEditingToken,
    closeEdit,

    // UI state
    compactMode,
    setCompactMode,
    showKeys,
    setShowKeys,
    resolvedTokenKeys,
    loadingTokenKeys,

    // Form state
    formApi,
    setFormApi,
    formInitValues,
    getFormValues,

    // Functions
    loadTokens,
    refresh,
    copyText,
    fetchTokenKey,
    toggleTokenVisibility,
    copyTokenKey,
    onConfigureProvider,
    onOpenLink,
    manageToken,
    searchTokens,
    sortToken,
    handlePageChange,
    handlePageSizeChange,
    rowSelection,
    handleRow,
    batchDeleteTokens,
    batchCopyTokens,
    syncPageData,

    // Translation
    t,
  };
};
