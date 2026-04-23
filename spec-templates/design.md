# <Project Name> — Design

**Date:** YYYY-MM-DD
**Status:** Draft | Active | Deprecated

## Architecture Overview

<High-level description of how the system is structured. Include a text diagram showing component relationships.>

```
Component A  ──┐
               ├──▶  Core Logic  ──▶  External Service
Component B  ──┘
```

## Project Structure

```
project/
├── main.go              # <Description>
├── cmd/                  # <Description>
├── pkg/                  # <Description>
└── ...
```

## Components

### Component: <Name>

- **Responsibility:** <What it does>
- **Interfaces:** <How other components interact with it>
- **Key Decisions:** <Why this approach was chosen over alternatives>

<!-- Repeat for each component -->

## Data Flow

1. <Input source> provides <what>
2. <Component> processes <how>
3. <Output> is delivered as <format>

## Dependencies

| Dependency | Purpose | Version |
|-----------|---------|---------|
| <package> | <why it's needed> | <version> |

## Authentication & Security

<!-- Remove if not applicable -->

<How credentials/secrets are handled. Auth flow description.>

## Error Handling Strategy

<!-- Remove if not applicable -->

<How errors propagate, what the user sees, retry logic if any.>

## Future Considerations

<!-- Optional: known extension points, areas for future growth -->

- <Potential enhancement and what it would require>
