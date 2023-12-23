#tensor_rank
let $graph = (cy graph-links-df --include_contracts)

let $links_nu = ($graph | dfr select particle_from particle_to | dfr into-nu)

print {
    'total links count': ($links_nu | length)
    'unique links count': ($links_nu | uniq-by particle_from particle_to | length)
    'repeated links unique count': ($links_nu | uniq-by particle_from particle_to --repeated | length)
}

$graph
| cy graph-neurons-stats
| dfr select [neuron, nickname, links_count, karma, first_link]
| dfr sort-by first_link
| dfr into-nu
| rename --column {index: neuron_id}
| move neuron_id --after neuron
| save -f '1_0_neuronid.csv'

'1_1_neurons_balances.yaml'
| if ($in | path exists) { rm $in }

print "Checking balances via cy"

open 1_0_neuronid.csv
| par-each {|i|
    cy tokens-balance-all $i.neuron --dont_convert_pools --routes to
    | cy tokens-sum
    | upsert neuron $i.neuron
    | save -a 1_1_neurons_balances.yaml
}

open 1_1_neurons_balances.yaml
| where denom in [milliampere millivolt]
| group-by neuron --to-table
| each {|i| $i | merge ($i.items | select denom amount | transpose -idr)}
| reject items
| rename neuron
| save -f 1_2_neuron_av_balances.csv

open 1_0_neuronid.csv
| join (open 1_2_neuron_av_balances.csv) neuron
| default 0 milliampere
| select neuron_id milliampere
| sort-by neuron_id
| save -f 1_3_neuronid_balance.csv


$graph
| dfr select particle_from particle_to
| dfr into-lazy
| dfr unique --subset [particle_from particle_to]
| dfr collect
| dfr with-column ( dfr arg-where ((dfr col particle_from) != '0') | dfr as link_id )
| dfr join ($graph) [particle_from particle_to] [particle_from particle_to]
| dfr select neuron link_id
| dfr to-csv '2_0_link_ids.csv'

open 2_0_link_ids.csv
| join (open 1_0_neuronid.csv) neuron
| select neuron_id link_id
| save -f 2_1_neuronid_linkid.csv

print "Running go script"

go run main.go
| save 3_1_go_output.txt -rf

open 3_1_go_output.txt -r
| lines
| last
| str replace 'Trust:  [' ''
| str trim -c ']'
| split row ' '
| into float
| wrap trust
| save -f 3_1_tensor_trust.csv


open 1_0_neuronid.csv
| merge (open 3_1_tensor_trust.csv)
| join (open 1_3_neuronid_balance.csv) neuron_id
| move trust milliampere --before links_count
| save -f 4_final.csv

print '4_final.csv is ready'

open 4_final.csv
