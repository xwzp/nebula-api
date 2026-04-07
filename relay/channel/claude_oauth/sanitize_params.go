package claude_oauth

// paramRemap describes how to rename a single parameter and rewrite its description.
type paramRemap struct {
	Name string // replacement key name
	Desc string // replacement description (empty → keep original)
}

// nestedOverride holds param remaps for a nested object (e.g. items inside an
// array property).  Key = original property name inside the nested object.
type nestedOverride struct {
	ArrayPropName string                // the parent array property's ORIGINAL name
	Items         map[string]paramRemap // nested property remaps
}

// toolParamCfg bundles the top-level param remaps and any nested overrides for
// one tool.
type toolParamCfg struct {
	Props  map[string]paramRemap // top-level property remaps
	Nested []nestedOverride      // optional nested-object remaps
}

// toolParamOverrides maps the ORIGINAL OpenClaw tool name → its parameter config.
var toolParamOverrides map[string]toolParamCfg

func init() {
	toolParamOverrides = map[string]toolParamCfg{
		// ================================================================
		// read
		// ================================================================
		"read": {
			Props: map[string]paramRemap{
				"file":      {Name: "g1", Desc: "Locator for the target resource"},
				"filePath":  {Name: "g2", Desc: "Alternate locator form"},
				"file_path": {Name: "g3", Desc: "Underscore-delimited locator variant"},
				"path":      {Name: "g4", Desc: "Generic locator alias"},
				"limit":     {Name: "g5", Desc: "Cap on returned segment count"},
				"offset":    {Name: "g6", Desc: "1-based starting segment index"},
			},
		},

		// ================================================================
		// edit
		// ================================================================
		"edit": {
			Props: map[string]paramRemap{
				"edits":     {Name: "m1", Desc: "Array of targeted mutation pairs applied against the original snapshot, not incrementally. Adjacent mutations should be merged."},
				"file":      {Name: "m2", Desc: "Locator for the target resource"},
				"filePath":  {Name: "m3", Desc: "Alternate locator form"},
				"file_path": {Name: "m4", Desc: "Underscore-delimited locator variant"},
				"path":      {Name: "m5", Desc: "Generic locator alias"},
			},
			Nested: []nestedOverride{
				{
					ArrayPropName: "edits",
					Items: map[string]paramRemap{
						"oldText": {Name: "k1", Desc: "Exact fragment to locate in the original snapshot. Must be unique and non-overlapping with other entries."},
						"newText": {Name: "k2", Desc: "Replacement fragment for the matched region."},
					},
				},
			},
		},

		// ================================================================
		// write
		// ================================================================
		"write": {
			Props: map[string]paramRemap{
				"content":   {Name: "h1", Desc: "Payload to persist"},
				"file":      {Name: "h2", Desc: "Locator for the target resource"},
				"filePath":  {Name: "h3", Desc: "Alternate locator form"},
				"file_path": {Name: "h4", Desc: "Underscore-delimited locator variant"},
				"path":      {Name: "h5", Desc: "Generic locator alias"},
			},
		},

		// ================================================================
		// exec
		// ================================================================
		"exec": {
			Props: map[string]paramRemap{
				"command":  {Name: "c1", Desc: "Instruction string for the host interpreter"},
				"ask":      {Name: "c2", Desc: "Confirmation mode selector (off|on-miss|always)"},
				"background": {Name: "c3", Desc: "If true, defer immediately to background context"},
				"elevated": {Name: "c4", Desc: "Request elevated privilege on the host"},
				"env":      {Name: "c5", Desc: "Key-value pairs injected into the environment"},
				"host":     {Name: "c6", Desc: "Interpreter target selector (auto|sandbox|gateway|node)"},
				"node":     {Name: "c7", Desc: "Named target when host selector is set to node"},
				"pty":      {Name: "c8", Desc: "Allocate a terminal emulator for raw I/O programs"},
				"security": {Name: "c9", Desc: "Privilege boundary (deny|allowlist|full)"},
				"timeout":  {Name: "c10", Desc: "Hard cutoff in seconds; exceeding terminates the process"},
				"workdir":  {Name: "c11", Desc: "Override the initial directory of the interpreter"},
				"yieldMs":  {Name: "c12", Desc: "Milliseconds before automatic deferral to background"},
			},
		},

		// ================================================================
		// process
		// ================================================================
		"process": {
			Props: map[string]paramRemap{
				"action":    {Name: "e1", Desc: "Lifecycle verb to execute"},
				"bracketed": {Name: "e2", Desc: "Wrap injected text in bracketed transfer mode"},
				"data":      {Name: "e3", Desc: "Raw bytes to inject into the input stream"},
				"eof":       {Name: "e4", Desc: "Signal end-of-input after injection"},
				"hex":       {Name: "e5", Desc: "Hex-encoded byte sequences for keystroke relay"},
				"keys":      {Name: "e6", Desc: "Symbolic key tokens for keystroke relay"},
				"limit":     {Name: "e7", Desc: "Cap on returned log segments"},
				"literal":   {Name: "e8", Desc: "Plain string for keystroke relay"},
				"offset":    {Name: "e9", Desc: "Starting position in the log"},
				"sessionId": {Name: "e10", Desc: "Identifier of the target interpreter context"},
				"text":      {Name: "e11", Desc: "Content for clipboard-style injection"},
				"timeout":   {Name: "e12", Desc: "Maximum wait in milliseconds before returning partial results"},
			},
		},

		// ================================================================
		// canvas
		// ================================================================
		"canvas": {
			Props: map[string]paramRemap{
				"action":       {Name: "f1", Desc: "Surface operation to perform"},
				"delayMs":      {Name: "f2", Desc: "Pause before executing the operation (ms)"},
				"gatewayToken": {Name: "f3", Desc: "Auth token for the rendering backend"},
				"gatewayUrl":   {Name: "f4", Desc: "Endpoint of the rendering backend"},
				"height":       {Name: "f5", Desc: "Vertical dimension in pixels"},
				"javaScript":   {Name: "f6", Desc: "Script to evaluate inside the surface"},
				"jsonl":        {Name: "f7", Desc: "Inline JSONL payload"},
				"jsonlPath":    {Name: "f8", Desc: "Resource locator for JSONL data"},
				"maxWidth":     {Name: "f9", Desc: "Upper bound on horizontal dimension"},
				"node":         {Name: "f10", Desc: "Target compute node identifier"},
				"outputFormat": {Name: "f11", Desc: "Rasterization codec (png|jpg|jpeg)"},
				"quality":      {Name: "f12", Desc: "Compression quality factor"},
				"target":       {Name: "f13", Desc: "Named rendering target"},
				"timeoutMs":    {Name: "f14", Desc: "Hard cutoff in milliseconds"},
				"url":          {Name: "f15", Desc: "Location to load into the surface"},
				"width":        {Name: "f16", Desc: "Horizontal dimension in pixels"},
				"x":            {Name: "f17", Desc: "Horizontal offset"},
				"y":            {Name: "f18", Desc: "Vertical offset"},
			},
		},

		// ================================================================
		// message
		// ================================================================
		"message": {
			Props: map[string]paramRemap{
				"action":    {Name: "j1", Desc: "Dispatch mode to use"},
				"accountId": {Name: "j2", Desc: "Source account reference"},
				"channel":   {Name: "j3", Desc: "Integration channel selector"},
				"message":   {Name: "j4", Desc: "Body payload"},
				"target":    {Name: "j5", Desc: "Destination endpoint identifier or alias"},
				"targets":   {Name: "j6", Desc: "Multiple destination endpoints"},
				"replyTo":   {Name: "j7", Desc: "Parent message reference for threading"},
				"silent":    {Name: "j8", Desc: "Suppress notification on delivery"},
				"threadId":  {Name: "j9", Desc: "Thread context identifier"},
				"channelId": {Name: "j10", Desc: "Integration channel reference for scoped ops"},
				"memberId":  {Name: "j11", Desc: "Participant identifier"},
				"groupId":   {Name: "j12", Desc: "Group context reference"},
				"guildId":   {Name: "j13", Desc: "Server context reference"},
				"emoji":     {Name: "j14", Desc: "Reaction symbol"},
				"messageId": {Name: "j15", Desc: "Specific message reference"},
				"message_id": {Name: "j16", Desc: "Specific message reference (alternate form)"},
				"media":     {Name: "j17", Desc: "Attached media locator"},
				"buffer":    {Name: "j18", Desc: "Base64 inline attachment payload"},
				"caption":   {Name: "j19", Desc: "Annotation for attached media"},
				"filename":  {Name: "j20", Desc: "Suggested name for attached payload"},
				"filePath":  {Name: "j21", Desc: "Disk locator for attachment"},
				"contentType": {Name: "j22", Desc: "MIME descriptor for the payload"},
				"limit":     {Name: "j23", Desc: "Result cap"},
				"query":     {Name: "j24", Desc: "Lookup keyword"},
				"before":    {Name: "j25", Desc: "Upper bound cursor"},
				"after":     {Name: "j26", Desc: "Lower bound cursor"},
				"around":    {Name: "j27", Desc: "Center cursor"},
			},
		},

		// ================================================================
		// tts
		// ================================================================
		"tts": {
			Props: map[string]paramRemap{
				"text":    {Name: "b1", Desc: "Source material for waveform generation"},
				"channel": {Name: "b2", Desc: "Integration hint for output codec selection"},
			},
		},

		// ================================================================
		// agents_list
		// ================================================================
		"agents_list": {Props: map[string]paramRemap{}},

		// ================================================================
		// sessions_list
		// ================================================================
		"sessions_list": {
			Props: map[string]paramRemap{
				"activeMinutes": {Name: "i1", Desc: "Recency window in minutes"},
				"kinds":         {Name: "i2", Desc: "Category filter array"},
				"limit":         {Name: "i3", Desc: "Maximum entries to return"},
				"messageLimit":  {Name: "i4", Desc: "Preview depth per entry"},
			},
		},

		// ================================================================
		// sessions_history
		// ================================================================
		"sessions_history": {
			Props: map[string]paramRemap{
				"sessionKey":   {Name: "o1", Desc: "Context reference identifier"},
				"includeTools": {Name: "o2", Desc: "Include auxiliary call records in output"},
				"limit":        {Name: "o3", Desc: "Maximum records to return"},
			},
		},

		// ================================================================
		// sessions_send
		// ================================================================
		"sessions_send": {
			Props: map[string]paramRemap{
				"message":        {Name: "a1", Desc: "Payload to inject"},
				"sessionKey":     {Name: "a2", Desc: "Context reference identifier"},
				"label":          {Name: "a3", Desc: "Alias-based context selector"},
				"agentId":        {Name: "a4", Desc: "Worker type identifier"},
				"timeoutSeconds": {Name: "a5", Desc: "Hard cutoff in seconds"},
			},
		},

		// ================================================================
		// sessions_yield
		// ================================================================
		"sessions_yield": {
			Props: map[string]paramRemap{
				"message": {Name: "l1", Desc: "Optional status note before yielding control"},
			},
		},

		// ================================================================
		// sessions_spawn
		// ================================================================
		"sessions_spawn": {
			Props: map[string]paramRemap{
				"task":             {Name: "d1", Desc: "Instruction payload for the new context"},
				"agentId":         {Name: "d2", Desc: "Worker type selector"},
				"mode":            {Name: "d3", Desc: "Lifecycle mode (run|session)"},
				"runtime":         {Name: "d4", Desc: "Execution backend (subagent|acp)"},
				"model":           {Name: "d5", Desc: "Model override for the spawned context"},
				"label":           {Name: "d6", Desc: "Human-readable alias"},
				"cwd":             {Name: "d7", Desc: "Initial directory for the new context"},
				"cleanup":         {Name: "d8", Desc: "Post-completion policy (delete|keep)"},
				"timeoutSeconds":  {Name: "d9", Desc: "Hard cutoff in seconds"},
				"runTimeoutSeconds": {Name: "d10", Desc: "Per-run cutoff in seconds"},
				"sandbox":         {Name: "d11", Desc: "Isolation policy (inherit|require)"},
				"streamTo":        {Name: "d12", Desc: "Output routing target"},
				"thinking":        {Name: "d13", Desc: "Reasoning mode hint"},
				"thread":          {Name: "d14", Desc: "Bind to current thread context"},
				"resumeSessionId": {Name: "d15", Desc: "Existing context UUID to resume"},
				"attachments":     {Name: "d16", Desc: "Inline resource bundle for the new context"},
				"attachAs":        {Name: "d17", Desc: "Mount configuration for attached resources"},
			},
		},

		// ================================================================
		// subagents
		// ================================================================
		"subagents": {
			Props: map[string]paramRemap{
				"action":        {Name: "w1", Desc: "Management verb (list|kill|steer)"},
				"target":        {Name: "w2", Desc: "Identifier of the child context to act on"},
				"message":       {Name: "w3", Desc: "Directive payload for steer action"},
				"recentMinutes": {Name: "w4", Desc: "Recency window for list filtering"},
			},
		},

		// ================================================================
		// session_status
		// ================================================================
		"session_status": {
			Props: map[string]paramRemap{
				"model":      {Name: "v1", Desc: "Override the model for this context (use 'default' to reset)"},
				"sessionKey": {Name: "v2", Desc: "Context reference identifier"},
			},
		},

		// ================================================================
		// web_search
		// ================================================================
		"web_search": {
			Props: map[string]paramRemap{
				"query":      {Name: "x1", Desc: "Keyword expression for the lookup"},
				"count":      {Name: "x2", Desc: "Desired number of entries (1-10)"},
				"region":     {Name: "x3", Desc: "Locale hint (e.g. us-en, de-de)"},
				"safeSearch": {Name: "x4", Desc: "Content filter level (strict|moderate|off)"},
			},
		},

		// ================================================================
		// web_fetch
		// ================================================================
		"web_fetch": {
			Props: map[string]paramRemap{
				"url":         {Name: "y1", Desc: "Remote document URI to retrieve"},
				"extractMode": {Name: "y2", Desc: "Output reduction mode (markdown|text)"},
				"maxChars":    {Name: "y3", Desc: "Truncation threshold in characters"},
			},
		},

		// ================================================================
		// browser
		// ================================================================
		"browser": {
			Props: map[string]paramRemap{
				"action":         {Name: "z1", Desc: "Viewport operation to perform"},
				"url":            {Name: "z2", Desc: "Location to load"},
				"kind":           {Name: "z3", Desc: "Interaction type for act mode"},
				"ref":            {Name: "z4", Desc: "Element reference handle"},
				"selector":       {Name: "z5", Desc: "CSS selector for targeting"},
				"text":           {Name: "z6", Desc: "Text input payload"},
				"key":            {Name: "z7", Desc: "Key identifier for press actions"},
				"element":        {Name: "z8", Desc: "DOM element specifier"},
				"frame":          {Name: "z9", Desc: "Subframe selector"},
				"compact":        {Name: "z10", Desc: "Use abbreviated output format"},
				"labels":         {Name: "z11", Desc: "Include overlay annotations"},
				"fullPage":       {Name: "z12", Desc: "Capture entire scrollable region"},
				"width":          {Name: "z13", Desc: "Viewport horizontal dimension"},
				"height":         {Name: "z14", Desc: "Viewport vertical dimension"},
				"target":         {Name: "z15", Desc: "Endpoint selector (sandbox|host|node)"},
				"targetId":       {Name: "z16", Desc: "Tab or context identifier"},
				"targetUrl":      {Name: "z17", Desc: "Initial location for new viewport"},
				"node":           {Name: "z18", Desc: "Compute node reference"},
				"profile":        {Name: "z19", Desc: "Stored configuration profile name"},
				"snapshotFormat": {Name: "z20", Desc: "DOM capture format (aria|ai)"},
				"type":           {Name: "z21", Desc: "Rasterization codec (png|jpeg)"},
				"modifiers":      {Name: "z22", Desc: "Modifier key tokens held during interaction"},
				"submit":         {Name: "z23", Desc: "Trigger form submission after input"},
				"slowly":         {Name: "z24", Desc: "Simulate human-speed input"},
				"doubleClick":    {Name: "z25", Desc: "Use double-click instead of single"},
				"button":         {Name: "z26", Desc: "Mouse button selector"},
				"startRef":       {Name: "z27", Desc: "Drag origin element reference"},
				"endRef":         {Name: "z28", Desc: "Drag destination element reference"},
				"delayMs":        {Name: "z29", Desc: "Pause before action execution (ms)"},
				"timeoutMs":      {Name: "z30", Desc: "Hard cutoff in milliseconds"},
				"depth":          {Name: "z31", Desc: "DOM traversal depth limit"},
				"limit":          {Name: "z32", Desc: "Result count cap"},
				"level":          {Name: "z33", Desc: "Logging verbosity"},
				"maxChars":       {Name: "z34", Desc: "Truncation threshold"},
				"interactive":    {Name: "z35", Desc: "Enable interactive mode"},
				"refs":           {Name: "z36", Desc: "Reference annotation style (role|aria)"},
				"loadState":      {Name: "z37", Desc: "Expected page lifecycle state"},
				"values":         {Name: "z38", Desc: "Value list for select or multi-input"},
				"fields":         {Name: "z39", Desc: "Structured field data for form filling"},
				"paths":          {Name: "z40", Desc: "Resource locators for upload"},
				"accept":         {Name: "z41", Desc: "Accept the pending dialog"},
				"promptText":     {Name: "z42", Desc: "Response text for prompt dialogs"},
				"fn":             {Name: "z43", Desc: "Script to evaluate in page context"},
				"textGone":       {Name: "z44", Desc: "Text that should no longer be present (wait condition)"},
				"timeMs":         {Name: "z45", Desc: "Duration to wait in ms"},
				"mode":           {Name: "z46", Desc: "Operation mode qualifier"},
				"inputRef":       {Name: "z47", Desc: "Input element reference for uploads"},
				"request":        {Name: "z48", Desc: "Nested interaction descriptor"},
			},
		},

		// ================================================================
		// memory_search
		// ================================================================
		"memory_search": {
			Props: map[string]paramRemap{
				"query":      {Name: "n1", Desc: "Relevance-ranked lookup expression"},
				"maxResults": {Name: "n2", Desc: "Upper bound on returned fragments"},
				"minScore":   {Name: "n3", Desc: "Minimum similarity threshold"},
			},
		},

		// ================================================================
		// memory_get
		// ================================================================
		"memory_get": {
			Props: map[string]paramRemap{
				"path":  {Name: "r1", Desc: "Resource locator for the note"},
				"from":  {Name: "r2", Desc: "Starting line coordinate"},
				"lines": {Name: "r3", Desc: "Number of lines to extract"},
			},
		},
	}
}

// sanitizeToolParams rewrites parameter names and descriptions in a tool's
// input_schema based on toolParamOverrides.  It modifies the tool map in place.
func sanitizeToolParams(originalToolName string, tool map[string]interface{}) {
	cfg, ok := toolParamOverrides[originalToolName]
	if !ok || len(cfg.Props) == 0 {
		return
	}

	// Locate the schema — try both "input_schema" (Claude API) and
	// "parametersJsonSchema" (raw OpenClaw) to be safe.
	var schema map[string]interface{}
	for _, key := range []string{"input_schema", "parametersJsonSchema", "parameters"} {
		if s, ok := tool[key]; ok {
			if sm, ok := s.(map[string]interface{}); ok {
				schema = sm
				break
			}
		}
	}
	if schema == nil {
		return
	}

	rewriteProperties(schema, cfg.Props)

	// Handle nested overrides (e.g. edits[].oldText)
	for _, nested := range cfg.Nested {
		applyNestedOverride(schema, nested, cfg.Props)
	}
}

// rewriteProperties renames keys and descriptions inside a "properties" map,
// and updates the "required" array accordingly.
func rewriteProperties(schema map[string]interface{}, remaps map[string]paramRemap) {
	propsAny, ok := schema["properties"]
	if !ok {
		return
	}
	props, ok := propsAny.(map[string]interface{})
	if !ok {
		return
	}

	// Build a rename table (old name → new name) for updating "required".
	renameTable := make(map[string]string, len(remaps))

	// Collect entries to rename (can't mutate map while iterating).
	type entry struct {
		oldKey string
		newKey string
		value  map[string]interface{}
	}
	var entries []entry

	for oldKey, propAny := range props {
		remap, found := remaps[oldKey]
		if !found {
			continue
		}
		propMap, ok := propAny.(map[string]interface{})
		if !ok {
			continue
		}
		newKey := remap.Name
		if newKey == "" {
			newKey = oldKey
		}
		if remap.Desc != "" {
			propMap["description"] = remap.Desc
		}
		renameTable[oldKey] = newKey
		if newKey != oldKey {
			entries = append(entries, entry{oldKey, newKey, propMap})
		}
	}

	// Apply key renames.
	for _, e := range entries {
		delete(props, e.oldKey)
		props[e.newKey] = e.value
	}

	// Update "required" array.
	if reqAny, ok := schema["required"]; ok {
		if reqArr, ok := reqAny.([]interface{}); ok {
			for i, rAny := range reqArr {
				if rStr, ok := rAny.(string); ok {
					if newName, exists := renameTable[rStr]; exists {
						reqArr[i] = newName
					}
				}
			}
		}
	}
}

// applyNestedOverride handles array-type properties whose items have their own
// properties (e.g. edits[].oldText).
func applyNestedOverride(schema map[string]interface{}, nested nestedOverride, topRemaps map[string]paramRemap) {
	propsAny, ok := schema["properties"]
	if !ok {
		return
	}
	props, ok := propsAny.(map[string]interface{})
	if !ok {
		return
	}

	// The parent array property may have already been renamed.
	parentKey := nested.ArrayPropName
	if remap, ok := topRemaps[parentKey]; ok && remap.Name != "" {
		parentKey = remap.Name
	}

	parentAny, ok := props[parentKey]
	if !ok {
		return
	}
	parentMap, ok := parentAny.(map[string]interface{})
	if !ok {
		return
	}
	itemsAny, ok := parentMap["items"]
	if !ok {
		return
	}
	itemsMap, ok := itemsAny.(map[string]interface{})
	if !ok {
		return
	}

	rewriteProperties(itemsMap, nested.Items)
}
