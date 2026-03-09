import { useState, useEffect } from 'react'
import {
  Collapsible,
  CollapsibleContent,
  CollapsibleTrigger,
} from '@/components/ui/collapsible'
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card'
import { Button } from '@/components/ui/button'
import { Input } from '@/components/ui/input'
import { Textarea } from '@/components/ui/textarea'
import { useObjective } from '@/hooks/useObjective'

interface ObjectivePanelProps {
  conversationId: string | null
}

export function ObjectivePanel({ conversationId }: ObjectivePanelProps) {
  const { objective, isLoading, error, upsert, isUpserting } = useObjective(conversationId)
  const [open, setOpen] = useState(false)
  const [title, setTitle] = useState('')
  const [goal, setGoal] = useState('')
  const [constraints, setConstraints] = useState('')
  const [toolsRequired, setToolsRequired] = useState('')
  const [successCriteria, setSuccessCriteria] = useState('')
  const [hasEdited, setHasEdited] = useState(false)

  useEffect(() => {
    if (!hasEdited) {
      if (objective) {
        setTitle(objective.title)
        setGoal(objective.goal)
        setConstraints(objective.constraints)
        setToolsRequired(objective.tools_required?.join(', ') ?? '')
        setSuccessCriteria(objective.success_criteria)
      } else if (!isLoading) {
        setTitle('')
        setGoal('')
        setConstraints('')
        setToolsRequired('')
        setSuccessCriteria('')
      }
    }
  }, [objective, isLoading, hasEdited])

  if (!conversationId) return null

  const onSave = async () => {
    await upsert({
      title,
      goal,
      constraints,
      tools_required: toolsRequired
        .split(',')
        .map((t) => t.trim())
        .filter(Boolean),
      success_criteria: successCriteria,
    })
    setHasEdited(false)
  }

  return (
    <Collapsible open={open} onOpenChange={setOpen}>
      <Card className="border-border/50">
        <CollapsibleTrigger asChild>
          <CardHeader className="cursor-pointer py-4 hover:bg-accent/50">
            <CardTitle className="text-base">
              Objective {objective ? `— ${objective.title || 'Untitled'}` : '(none)'}
            </CardTitle>
          </CardHeader>
        </CollapsibleTrigger>
        <CollapsibleContent>
          <CardContent className="space-y-4 pt-0">
            {error && <p className="text-sm text-destructive">Failed to load objective.</p>}
            {isLoading ? (
              <p className="text-sm text-muted-foreground">Loading...</p>
            ) : (
              <>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Title</label>
                  <Input
                    value={title}
                    onChange={(e) => {
                      setTitle(e.target.value)
                      setHasEdited(true)
                    }}
                    placeholder="Short name of the skill or goal"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Goal</label>
                  <Textarea
                    value={goal}
                    onChange={(e) => {
                      setGoal(e.target.value)
                      setHasEdited(true)
                    }}
                    placeholder="What you want to achieve"
                    rows={3}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Constraints</label>
                  <Textarea
                    value={constraints}
                    onChange={(e) => {
                      setConstraints(e.target.value)
                      setHasEdited(true)
                    }}
                    placeholder="Time, budget, safety, environment limits"
                    rows={2}
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Tools required (comma-separated)</label>
                  <Input
                    value={toolsRequired}
                    onChange={(e) => {
                      setToolsRequired(e.target.value)
                      setHasEdited(true)
                    }}
                    placeholder="filesystem, git, summarize"
                  />
                </div>
                <div className="space-y-2">
                  <label className="text-sm font-medium">Success criteria</label>
                  <Textarea
                    value={successCriteria}
                    onChange={(e) => {
                      setSuccessCriteria(e.target.value)
                      setHasEdited(true)
                    }}
                    placeholder="What counts as success"
                    rows={2}
                  />
                </div>
                {hasEdited && (
                  <Button onClick={onSave} disabled={isUpserting}>
                    {isUpserting ? 'Saving...' : 'Save'}
                  </Button>
                )}
              </>
            )}
          </CardContent>
        </CollapsibleContent>
      </Card>
    </Collapsible>
  )
}
