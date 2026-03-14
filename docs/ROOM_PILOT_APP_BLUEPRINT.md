# Room Pilot

`Room Pilot` is a flagship MCP app for furnishing real rooms inside OpenADE.

It should feel like a live design-and-buying war room:

1. describe a room
2. define budget and taste
3. source real products
4. build bundles
5. compare options
6. inspect fit and delivery risks
7. stage an approval-ready order plan

This is not a replacement for the current OpenADE loop. It is a strong app built on top of:

`chat -> objective -> task -> run`

## Why This Is A Better Flagship App

`Room Pilot` is more impressive than a generic capability builder because it is:

- visual
- interactive
- budget-constrained
- taste-sensitive
- tool-heavy
- immediately understandable
- easy to demo

It gives the model a real job:

- interpret a room
- extract constraints
- form a style direction
- source products
- build coherent bundles
- explain tradeoffs
- revise fast

## The Core User Loop

The real loop should be:

1. `Brief`
2. `Plan`
3. `Source`
4. `Bundle`
5. `Compare`
6. `Approve`

In OpenADE terms:

1. `Chat` captures the room goal
2. `Objective` stores the furnishing brief
3. `Tasks` package reusable sourcing and planning behaviors
4. `Runs` execute bundle generation and comparison
5. `MCP App` becomes the interactive operating surface

## What The User Can Ask For

Examples:

- furnish a studio apartment under $2,500
- make this living room warmer and less cluttered
- replace everything that ships slower than 2 weeks
- keep the sofa and redesign the rest
- make it renter-friendly
- make it kid-safe
- optimize for pets
- show 3 room packages at different price levels

## How It Fits The Current OpenADE UI

Use the current shell first. Add `Room Pilot` as a flagship app, not a separate product.

### 1. Chat

Use `Chat` to collect the messy real-world brief:

- room type
- room dimensions
- budget
- style words
- must-keep items
- avoid items
- delivery constraints
- renter / pet / child constraints

The model should ask follow-up questions until the room is usable.

### 2. Objective

Use the existing objective panel as the project brief.

Recommended fields for a furnishing session:

- `Title`
- `Goal`
- `Constraints`
- `Tools required`
- `Success criteria`

Example:

- `Title`: Furnish Brooklyn Studio
- `Goal`: Build a coherent living + sleeping setup under $3,000
- `Constraints`: renter-friendly, warm minimal, narrow walkways, cat-safe
- `Tools required`: search, browser, product extractor, budget checker
- `Success criteria`: 2-3 viable bundles with real products and clear tradeoffs

### 3. Tasks

Use tasks to package the reusable pieces of behavior behind the app.

`Room Pilot` should initially be powered by reusable tasks, not by one giant monolith.

### 4. Runs

Use runs to inspect:

- product bundles
- comparison output
- fit warnings
- reflection
- what worked and what should change

### 5. Room Pilot App Surface

After the core tasks work, expose them through a richer app page inside OpenADE.

This should eventually become the first serious interactive MCP app host surface.

## The First 5 Tasks

These are the minimum useful tasks to build first.

### 1. Room Constraint Extractor

Purpose:
- turn messy user input into a structured room brief

Output:
- room dimensions
- required furniture types
- hard constraints
- soft preferences
- risk flags

### 2. Style Brief Builder

Purpose:
- turn taste, adjectives, and inspiration into a clear design direction

Output:
- style summary
- color/material guidance
- do / avoid list
- style keywords for sourcing

### 3. Furniture Bundle Builder

Purpose:
- propose coherent product bundles under budget and fit constraints

Output:
- bundle name
- products
- prices
- rationale
- tradeoffs

### 4. Bundle Comparison Reviewer

Purpose:
- compare multiple room packages and explain why one is better

Output:
- side-by-side comparison
- budget delta
- fit delta
- style delta
- recommended winner

### 5. Order Stager

Purpose:
- convert a chosen bundle into an approval-ready action sheet

Output:
- vendor
- product links
- SKU or product name
- quantity
- subtotal
- shipping caveats
- manual approval checklist

## The First Real MCP App Surface

The first version of the app should have 6 sections.

### Room Intake

Collect:

- room type
- dimensions
- budget
- style words
- must-keep furniture
- avoid list
- special constraints

This can start as a structured form backed by the objective and extractor task.

### Room Canvas

A simple scaled room layout:

- walls
- windows
- doors
- furniture footprints
- walkway hints

Version 1 does not need to be fancy. Rectangles on a grid are enough.

### Style Board

Show:

- style brief
- color/material tags
- “warmer / cooler / more premium / smaller-space friendly” controls

This should make revision feel immediate.

### Catalog Explorer

Show sourced items from MCP-backed tools:

- image
- title
- vendor
- price
- dimensions
- lead time
- notes

### Bundle Builder

Assemble complete room packages:

- seating
- table
- rug
- storage
- lighting
- accent items

The user should be able to swap one item without rebuilding everything.

### Approval Console

Show:

- final links
- vendor
- quantity
- budget status
- warnings
- manual next step

Important:
- no auto-purchase in v1
- no hidden pricing
- no silent substitutions

## MCP / Tool Requirements

This app needs a real starter toolkit.

### Minimum Tooling

1. search / fetch
2. structured product extraction
3. browser automation
4. filesystem save/export
5. memory
6. secret handling

### Best Early MCP Categories

1. web search
2. browser control
3. product page fetch / scrape
4. structured table or JSON output
5. email follow-up later

### High-Value Later Integrations

1. 1Password CLI or provider
2. AgentMail
3. browser-agent style purchase flow
4. retailer-specific connectors

## Safety Rules

These are product rules, not legal boilerplate.

1. never purchase automatically
2. always show vendor and price
3. always show dimensions when available
4. always show budget overrun clearly
5. always surface fit risks
6. require explicit approval for any cart or checkout action

## Likely New Data In The Backend

Keep this loose for now. These are likely additions, not final contracts.

### Furniture Project

A project anchored to a conversation or objective, including:

- room name
- room type
- budget
- style brief
- room dimensions
- constraints

### Product Snapshot

A stored product reference so runs remain inspectable later:

- title
- vendor
- URL
- image
- price
- dimensions
- lead time
- extracted notes

### Bundle

A saved proposed room package:

- bundle name
- included items
- subtotal
- warnings
- why it was recommended

### Comparison

A stored comparison between multiple bundles.

## Likely New API Surfaces

Again: loose and practical, not final contracts.

### Room Project APIs

- create/update room project
- attach room project to a conversation
- fetch room project state

### Bundle APIs

- generate bundle
- list bundles
- compare bundles
- stage selected bundle

### Product Snapshot APIs

- save sourced products from a run
- fetch products by project or bundle

## UI Build Order

### Milestone 1

Build the capability using the system that already exists:

1. objective for the room
2. the five reusable tasks
3. run inspection
4. manual MCP tool use for sourcing

Definition of done:

- a user can produce 2-3 real furnishing bundles from chat + task runs

### Milestone 2

Add a first-party `Room Pilot` page inside OpenADE:

1. room intake
2. style board
3. bundle list
4. comparison view
5. approval console

Definition of done:

- a user can stay inside one app surface instead of jumping between raw task panels

### Milestone 3

Add richer MCP interaction:

1. product sourcing directly from the app
2. saved snapshots
3. bundle comparison
4. browser-assisted link validation

Definition of done:

- the app feels like a real shopping/design workstation

### Milestone 4

Add serious operating features:

1. secrets support
2. quote / email generation
3. manual cart staging
4. saved room templates
5. remixable room styles

Definition of done:

- users can come back repeatedly and reuse prior capability patterns

## Why This Matters For OpenADE

`Room Pilot` is not just a furniture app.

It proves the real direction of OpenADE:

- conversational objective setting
- reusable capability tasks
- tool-backed execution
- rich interactive MCP app surfaces
- real-world decisions with budget and approval loops

If this app works, OpenADE starts to feel like an actual agentic environment instead of just a chat shell with settings.
