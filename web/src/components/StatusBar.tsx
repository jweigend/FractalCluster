interface Props {
  completed: number;
  total: number;
  calculating: boolean;
}

export function StatusBar({ completed, total, calculating }: Props) {
  const pct = total > 0 ? Math.round((completed / total) * 100) : 0;

  return (
    <div style={styles.bar}>
      <div style={styles.info}>
        {calculating
          ? `Computing: ${completed} / ${total} blocks (${pct}%)`
          : total > 0
          ? `Done: ${total} blocks`
          : "Ready"}
      </div>
      <div style={styles.track}>
        <div style={{ ...styles.fill, width: `${pct}%` }} />
      </div>
    </div>
  );
}

const styles: Record<string, React.CSSProperties> = {
  bar: {
    display: "flex",
    alignItems: "center",
    gap: 12,
    padding: "6px 16px",
    background: "#1a1a2e",
    borderTop: "1px solid #333",
    fontSize: 12,
    color: "#aaa",
  },
  info: { whiteSpace: "nowrap" },
  track: {
    flex: 1,
    height: 8,
    background: "#0d0d1a",
    borderRadius: 4,
    overflow: "hidden",
  },
  fill: {
    height: "100%",
    background: "#2a6",
    transition: "width 0.15s ease",
    borderRadius: 4,
  },
};
