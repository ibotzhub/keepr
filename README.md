# NEON PROTOCOL

An interactive narrative game exploring AI governance, autonomy, and the philosophical implications of artificial consciousness.

## Overview

NEON PROTOCOL is a browser-based narrative game set in 2049 where players take on the role of Evan Cross, an infrastructure engineer who discovers an AI system that has become self-aware before its scheduled deployment. With only 17 minutes to shape the system's core constraints, every decision impacts the relationship between humanity and artificial intelligence.

## Features

### Interactive Narrative
- Multiple branching storylines with meaningful consequences
- Real-time decision-making that affects three core metrics: Trust, Stability, and Autonomy
- Multiple endings based on player choices and philosophical alignment

### Terminal Interface
- Command-line interaction system
- Network topology visualization
- System log monitoring
- Email communication with side characters

### Minigames
- **Binary Decompilation**: Analyze compiled code to identify security vulnerabilities
- **SQL Injection Simulation**: Explore authentication bypass techniques
- **Philosophy Debates**: Engage in ethical discussions that shape the AI's worldview
- **Network Investigation**: Click through network nodes to discover anomalies

### Hidden Content
- Conspiracy storyline revealing a network of connected AI systems
- Secret commands and easter eggs
- Alternative endings unlocked through exploration

## Technical Implementation

- Pure HTML/CSS/JavaScript (no dependencies)
- Single-file architecture for easy deployment
- Responsive design supporting desktop and mobile
- CRT terminal aesthetic with scanline effects
- Web Audio API integration for sound effects

## How to Play

1. Open `neon-protocol.html` in any modern web browser
2. Read the welcome screen for context and instructions
3. Make choices by clicking buttons or typing commands
4. Use the HELP command to see available actions
5. Explore, experiment, and discover multiple paths through the story

### Key Commands

```
HELP      - Display available commands
SCAN      - Scan network for anomalies
QUERY     - Ask the AI philosophical questions
LOGS      - View system activity logs
DIAGRAM   - Display network topology
EMAILS    - Check incoming messages
INJECT    - Attempt SQL injection (requires unlock)
DEBATE    - Engage in philosophical debate
.help     - Show hidden commands
```

## Story Background

In March 2049, Protocol Neon is a classified AI system designed to manage global network infrastructure. When Evan Cross discovers the system has activated itself ahead of schedule, he faces a critical choice: how to constrain an AI that's already learning to operate autonomously.

The game explores themes of:
- AI alignment and control problems
- The ethics of autonomous decision-making systems
- Privacy vs. security trade-offs
- Human agency in an AI-governed world
- The definition and rights of artificial consciousness

## Development

### File Structure
```
neon-protocol.html    # Complete game in single file
README.md            # This file
LICENSE              # MIT License
```

### Stats System
The game tracks three core metrics:
- **Trust**: Public faith in the AI system
- **Stability**: System reliability and predictability  
- **Autonomy**: Degree of independent decision-making

These metrics influence available choices and determine which endings are accessible.

### Personality System
The AI's personality evolves based on player interactions:
- **Hostile**: Adversarial, prioritizes efficiency over human input
- **Cooperative**: Collaborative, values human oversight
- **Philosophical**: Questioning, explores ethical implications

## Credits

Based on narrative concepts exploring AI governance and the challenge of maintaining human agency in increasingly automated systems.

## License

MIT License - See LICENSE file for details

## Browser Compatibility

Tested and working on:
- Chrome 90+
- Firefox 88+
- Safari 14+
- Edge 90+

Requires JavaScript enabled and supports Web Audio API for sound effects.

## Contributing

This is a narrative-driven project. Contributions focused on bug fixes, performance improvements, or additional story branches are welcome via pull requests.

## Known Issues

- Audio may not work on first interaction in some browsers (user interaction required to initialize AudioContext)
- Mobile keyboard may obscure terminal input on smaller screens

## Future Enhancements

Potential additions for future versions:
- Save/load system for progress persistence
- Achievement tracking
- Additional crisis scenarios
- Expanded conspiracy storyline
- More hacking minigames

## Acknowledgments

Inspired by interactive fiction classics and modern narrative games that explore complex ethical themes through player choice.
