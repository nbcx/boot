package boot

type Default struct{}

// Run: Typically the actual work function. Most commands will only implement this.
func (p *Default) Run(cmd Commander, args []string) {

}
